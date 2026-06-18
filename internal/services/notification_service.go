package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type NotificationService interface {
	// Create and send notifications
	SendNotification(ctx context.Context, userID string, notif *NotificationInput) (*models.Notification, error)
	SendBulkNotifications(ctx context.Context, userIDs []string, notif *NotificationInput) error

	// Get notifications
	GetNotifications(ctx context.Context, userID string, params *models.NotificationListParams) (*models.NotificationListResponse, error)
	GetUnreadNotifications(ctx context.Context, userID string, limit int) ([]*models.Notification, error)
	GetNotification(ctx context.Context, id, userID string) (*models.Notification, error)

	// Notification actions
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkMultipleAsRead(ctx context.Context, ids []string, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Archive(ctx context.Context, id, userID string) error
	Delete(ctx context.Context, id, userID string) error
	DeleteAll(ctx context.Context, userID string) error

	// Counts
	GetUnreadCount(ctx context.Context, userID string) (*models.NotificationCounts, error)

	// Preferences
	GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, userID string, req *UpdatePreferencesRequest) (*models.NotificationPreferences, error)

	// Cleanup
	DeleteOldNotifications(ctx context.Context, days int) error
	StartCleanupRoutine(ctx context.Context, interval time.Duration, retentionDays int)

	// Deferred delivery
	StartDeferredDeliveryRoutine(ctx context.Context, interval time.Duration)

	// Event triggers (called by other services)
	NotifyApplicationStatusChange(ctx context.Context, applicationID, employeeID, status string) error
	NotifyNewApplication(ctx context.Context, jobID, employerID, employeeID string) error
	NotifyInterviewScheduled(ctx context.Context, applicationID, employeeID, employerID string, interviewDate time.Time) error
	NotifyNewJobMatch(ctx context.Context, employeeID string, jobID string, matchScore int) error
	NotifyJobAlert(ctx context.Context, employeeID string, jobCount int) error
	NotifyMessageReceived(ctx context.Context, userID string, fromUserID string, message string) error

	// Job event triggers
	NotifyJobCreated(ctx context.Context, employerID string, jobID string, jobTitle string) error
	NotifyJobUpdated(ctx context.Context, employerID string, jobID string, jobTitle string) error
	NotifyJobClosed(ctx context.Context, employerID string, jobID string, jobTitle string) error
	NotifyJobDeleted(ctx context.Context, employerID string, jobID string, jobTitle string) error

	// Withdrawal trigger
	NotifyApplicationWithdrawn(ctx context.Context, employerID string, jobID string, employeeName string) error

	// Admin action triggers
	NotifyJobApproved(ctx context.Context, employerID string, jobID string, jobTitle string) error
	NotifyJobRejected(ctx context.Context, employerID string, jobID string, jobTitle string, reason string) error
	NotifyContentApproved(ctx context.Context, reporterID string, entityType string, entityID string) error
	NotifyContentRejected(ctx context.Context, reporterID string, entityType string, entityID string, reason string) error
	NotifyCompanyVerified(ctx context.Context, companyID string, companyName string) error
	NotifyCompanyRejected(ctx context.Context, companyID string, companyName string, reason string) error
	NotifyUserSuspended(ctx context.Context, userID string) error
	NotifyUserActivated(ctx context.Context, userID string) error
	NotifyAdminNewReport(ctx context.Context, adminID string, entityType string, entityID string) error
	NotifyAdminNewVerificationRequest(ctx context.Context, adminID string, companyName string) error
}

type NotificationInput struct {
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	ActionURL string                 `json:"action_url"`
	Icon      string                 `json:"icon"`
	Priority  string                 `json:"priority"`
}

type UpdatePreferencesRequest struct {
	PushEnabled        *bool   `json:"push_enabled"`
	PushSound          *bool   `json:"push_sound"`
	EmailEnabled       *bool   `json:"email_enabled"`
	EmailDigest        *bool   `json:"email_digest"`
	InAppEnabled       *bool   `json:"in_app_enabled"`
	ApplicationUpdates *bool   `json:"application_updates"`
	JobAlerts          *bool   `json:"job_alerts"`
	InterviewReminders *bool   `json:"interview_reminders"`
	Messages           *bool   `json:"messages"`
	SystemAlerts       *bool   `json:"system_alerts"`
	Marketing          *bool   `json:"marketing"`
	DigestFrequency    *string `json:"digest_frequency"`
	QuietHoursEnabled  *bool   `json:"quiet_hours_enabled"`
	QuietStartHour     *int    `json:"quiet_start_hour"`
	QuietEndHour       *int    `json:"quiet_end_hour"`
	QuietTimezone      *string `json:"quiet_timezone"`
}

type NotificationServiceImpl struct {
	cfg              *config.Config
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	websocketHub     *WebSocketHub
	emailService     EmailService
}

func NewNotificationService(
	cfg *config.Config,
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	emailService EmailService,
) NotificationService {
	svc := &NotificationServiceImpl{
		cfg:              cfg,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		websocketHub:     GetHub(),
		emailService:     emailService,
	}
	svc.StartDeferredDeliveryRoutine(context.Background(), 5*time.Minute)
	return svc
}

func (s *NotificationServiceImpl) SendNotification(ctx context.Context, userID string, input *NotificationInput) (*models.Notification, error) {
	// Check user preferences
	prefs, err := s.notificationRepo.GetPreferences(ctx, userID)
	if err != nil {
		// Create default preferences
		prefs = &models.NotificationPreferences{
			UserID: userID,
		}
		s.notificationRepo.CreatePreferences(ctx, prefs)
	}

	// Check if notification type is enabled
	if !s.isNotificationTypeEnabled(prefs, input.Type) {
		return nil, nil
	}

	// Check quiet hours — defer instead of dropping
	if s.isInQuietHours(prefs) {
		return s.deferNotification(ctx, userID, input, prefs)
	}

	// Create notification record
	metadataJSON, _ := json.Marshal(input.Metadata)
	var metadata models.JSONMap
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &metadata)
	}
	if metadata == nil {
		metadata = make(models.JSONMap)
	}

	notification := &models.Notification{
		UserID:    userID,
		Type:      input.Type,
		Title:     input.Title,
		Content:   input.Content,
		Metadata:  metadata,
		ActionURL: input.ActionURL,
		Icon:      input.Icon,
		Priority:  input.Priority,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	now := time.Now()
	notification.CreatedAt = now
	notification.DeliveredAt = &now

	if notification.Priority == "" {
		notification.Priority = "normal"
	}
	if notification.Icon == "" {
		notification.Icon = s.getIconForType(input.Type)
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send real-time via WebSocket if in-app enabled
	if prefs.InAppEnabled {
		s.websocketHub.SendToUser(userID, models.WebSocketMessage{
			Type:      "notification",
			Data:      notification,
			Timestamp: time.Now(),
		})
	}

	// Send email if enabled and high priority
	if prefs.EmailEnabled && input.Priority == "high" {
		go s.sendEmailNotification(notification, userID)
	}

	return notification, nil
}

func (s *NotificationServiceImpl) SendBulkNotifications(ctx context.Context, userIDs []string, input *NotificationInput) error {
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	for _, userID := range userIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(uid string) {
			defer wg.Done()
			defer func() { <-sem }()
			s.SendNotification(context.Background(), uid, input)
		}(userID)
	}
	wg.Wait()
	return nil
}

func (s *NotificationServiceImpl) GetNotifications(ctx context.Context, userID string, params *models.NotificationListParams) (*models.NotificationListResponse, error) {
	if params == nil {
		params = &models.NotificationListParams{}
	}
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var notifications []*models.Notification
	var total int64
	var err error

	switch {
	case params.UnreadOnly:
		notifications, total, err = s.notificationRepo.ListUnreadFiltered(ctx, userID, params)
	case params.Type != "":
		notifications, total, err = s.notificationRepo.ListByType(ctx, userID, params.Type, params.Page, params.Limit)
	default:
		notifications, total, err = s.notificationRepo.ListByUser(ctx, userID, params.Page, params.Limit)
	}
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		unreadCount = 0
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &models.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          params.Page,
		Limit:         params.Limit,
		TotalPages:    totalPages,
	}, nil
}

func (s *NotificationServiceImpl) GetUnreadNotifications(ctx context.Context, userID string, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.ListUnread(ctx, userID, limit)
}

func (s *NotificationServiceImpl) GetNotification(ctx context.Context, id, userID string) (*models.Notification, error) {
	return s.notificationRepo.GetByID(ctx, id, userID)
}

func (s *NotificationServiceImpl) MarkAsRead(ctx context.Context, id, userID string) error {
	return s.notificationRepo.MarkAsRead(ctx, id, userID)
}

func (s *NotificationServiceImpl) MarkMultipleAsRead(ctx context.Context, ids []string, userID string) error {
	return s.notificationRepo.MarkMultipleAsRead(ctx, ids, userID)
}

func (s *NotificationServiceImpl) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationServiceImpl) Archive(ctx context.Context, id, userID string) error {
	return s.notificationRepo.Archive(ctx, id, userID)
}

func (s *NotificationServiceImpl) Delete(ctx context.Context, id, userID string) error {
	return s.notificationRepo.Delete(ctx, id, userID)
}

func (s *NotificationServiceImpl) DeleteAll(ctx context.Context, userID string) error {
	return s.notificationRepo.DeleteAll(ctx, userID)
}

func (s *NotificationServiceImpl) DeleteOldNotifications(ctx context.Context, days int) error {
	return s.notificationRepo.DeleteOldNotifications(ctx, days)
}

func (s *NotificationServiceImpl) StartCleanupRoutine(ctx context.Context, interval time.Duration, retentionDays int) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.DeleteOldNotifications(ctx, retentionDays); err != nil {
					fmt.Printf("[NotificationService] Cleanup error: %v\n", err)
				}
			}
		}
	}()
}

func (s *NotificationServiceImpl) GetUnreadCount(ctx context.Context, userID string) (*models.NotificationCounts, error) {
	total, err := s.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	byType, err := s.notificationRepo.GetUnreadCountByType(ctx, userID)
	if err != nil {
		byType = make(map[string]int)
	}

	highPriority, err := s.notificationRepo.GetHighPriorityUnreadCount(ctx, userID)
	if err != nil {
		highPriority = 0
	}

	return &models.NotificationCounts{
		TotalUnread:       total,
		ByType:            byType,
		HighPriorityCount: highPriority,
	}, nil
}

func (s *NotificationServiceImpl) GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error) {
	prefs, err := s.notificationRepo.GetPreferences(ctx, userID)
	if err != nil {
		// Create default preferences
		defaultPrefs := &models.NotificationPreferences{
			UserID:             userID,
			PushEnabled:        true,
			PushSound:          true,
			EmailEnabled:       true,
			InAppEnabled:       true,
			ApplicationUpdates: true,
			JobAlerts:          true,
			InterviewReminders: true,
			Messages:           true,
			SystemAlerts:       true,
			Marketing:          false,
		}
		if err := s.notificationRepo.CreatePreferences(ctx, defaultPrefs); err != nil {
			return nil, err
		}
		return defaultPrefs, nil
	}
	return prefs, nil
}

func (s *NotificationServiceImpl) UpdatePreferences(ctx context.Context, userID string, req *UpdatePreferencesRequest) (*models.NotificationPreferences, error) {
	updates := make(map[string]interface{})

	if req.PushEnabled != nil {
		updates["push_enabled"] = *req.PushEnabled
	}
	if req.PushSound != nil {
		updates["push_sound"] = *req.PushSound
	}
	if req.EmailEnabled != nil {
		updates["email_enabled"] = *req.EmailEnabled
	}
	if req.EmailDigest != nil {
		updates["email_digest"] = *req.EmailDigest
	}
	if req.InAppEnabled != nil {
		updates["in_app_enabled"] = *req.InAppEnabled
	}
	if req.ApplicationUpdates != nil {
		updates["application_updates"] = *req.ApplicationUpdates
	}
	if req.JobAlerts != nil {
		updates["job_alerts"] = *req.JobAlerts
	}
	if req.InterviewReminders != nil {
		updates["interview_reminders"] = *req.InterviewReminders
	}
	if req.Messages != nil {
		updates["messages"] = *req.Messages
	}
	if req.SystemAlerts != nil {
		updates["system_alerts"] = *req.SystemAlerts
	}
	if req.Marketing != nil {
		updates["marketing"] = *req.Marketing
	}
	if req.DigestFrequency != nil {
		updates["digest_frequency"] = *req.DigestFrequency
	}
	if req.QuietHoursEnabled != nil {
		updates["quiet_hours_enabled"] = *req.QuietHoursEnabled
	}
	if req.QuietStartHour != nil {
		updates["quiet_start_hour"] = *req.QuietStartHour
	}
	if req.QuietEndHour != nil {
		updates["quiet_end_hour"] = *req.QuietEndHour
	}
	if req.QuietTimezone != nil {
		updates["quiet_timezone"] = *req.QuietTimezone
	}

	if len(updates) > 0 {
		if err := s.notificationRepo.UpdatePreferences(ctx, userID, updates); err != nil {
			return nil, err
		}
	}

	return s.GetPreferences(ctx, userID)
}

// Event triggers

func (s *NotificationServiceImpl) NotifyApplicationStatusChange(ctx context.Context, applicationID, employeeID, status string) error {
	statusMessages := map[string]string{
		"shortlisted": "Your application has been shortlisted!",
		"rejected":    "Your application was not selected",
		"hired":       "Congratulations! You've been hired!",
		"interview":   "Interview scheduled for your application",
		"withdrawn":   "You have withdrawn your application",
	}

	message, ok := statusMessages[status]
	if !ok {
		message = "Your application status has been updated"
	}

	input := &NotificationInput{
		Type:      "application",
		Title:     "Application Status Update",
		Content:   message,
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employee/applications/%s", applicationID),
		Icon:      "briefcase",
	}

	_, err := s.SendNotification(ctx, employeeID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyInterviewScheduled(ctx context.Context, applicationID, employeeID, employerID string, interviewDate time.Time) error {
	input := &NotificationInput{
		Type:      "interview",
		Title:     "Interview Scheduled",
		Content:   fmt.Sprintf("Your interview has been scheduled for %s", interviewDate.Format("Monday, January 2, 2006 at 3:04 PM")),
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employee/applications/%s", applicationID),
		Icon:      "calendar",
	}

	_, err := s.SendNotification(ctx, employeeID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyNewApplication(ctx context.Context, jobID, employerID, employeeID string) error {
	input := &NotificationInput{
		Type:      "application",
		Title:     "New Application Received",
		Content:   "A new candidate has applied for your job posting",
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employer/jobs/%s/applications", jobID),
		Icon:      "user-plus",
	}

	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyNewJobMatch(ctx context.Context, employeeID string, jobID string, matchScore int) error {
	input := &NotificationInput{
		Type:      "alert",
		Title:     "New Job Match!",
		Content:   fmt.Sprintf("We found a job that matches your profile with %d%% compatibility", matchScore),
		Priority:  "normal",
		ActionURL: fmt.Sprintf("/jobs/%s", jobID),
		Icon:      "star",
	}

	_, err := s.SendNotification(ctx, employeeID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyJobAlert(ctx context.Context, employeeID string, jobCount int) error {
	input := &NotificationInput{
		Type:      "alert",
		Title:     "New Jobs Available",
		Content:   fmt.Sprintf("%d new jobs matching your preferences have been posted", jobCount),
		Priority:  "low",
		ActionURL: "/jobs",
		Icon:      "bell",
	}

	_, err := s.SendNotification(ctx, employeeID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyMessageReceived(ctx context.Context, userID string, fromUserID string, message string) error {
	messagesPath := "/employee/messages"
	if user, err := s.userRepo.GetByID(ctx, userID); err == nil && user != nil && user.Role == "employer" {
		messagesPath = "/employer/messages"
	}

	input := &NotificationInput{
		Type:      "message",
		Title:     "New Message",
		Content:   message,
		Priority:  "normal",
		ActionURL: messagesPath,
		Icon:      "message",
	}

	_, err := s.SendNotification(ctx, userID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyJobCreated(ctx context.Context, employerID string, jobID string, jobTitle string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Job Posted",
		Content:   fmt.Sprintf("Your job \"%s\" has been posted successfully", jobTitle),
		Priority:  "normal",
		ActionURL: fmt.Sprintf("/employer/jobs/%s", jobID),
		Icon:      "briefcase",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyJobUpdated(ctx context.Context, employerID string, jobID string, jobTitle string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Job Updated",
		Content:   fmt.Sprintf("Your job \"%s\" has been updated", jobTitle),
		Priority:  "normal",
		ActionURL: fmt.Sprintf("/employer/jobs/%s", jobID),
		Icon:      "edit",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyJobClosed(ctx context.Context, employerID string, jobID string, jobTitle string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Job Closed",
		Content:   fmt.Sprintf("Your job \"%s\" has been closed", jobTitle),
		Priority:  "normal",
		ActionURL: fmt.Sprintf("/employer/jobs/%s", jobID),
		Icon:      "lock",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyJobDeleted(ctx context.Context, employerID string, jobID string, jobTitle string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Job Deleted",
		Content:   fmt.Sprintf("Your job \"%s\" has been deleted", jobTitle),
		Priority:  "normal",
		ActionURL: "/employer/jobs",
		Icon:      "trash",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyApplicationWithdrawn(ctx context.Context, employerID string, jobID string, employeeName string) error {
	input := &NotificationInput{
		Type:      "application",
		Title:     "Application Withdrawn",
		Content:   fmt.Sprintf("%s has withdrawn their application", employeeName),
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employer/jobs/%s/applications", jobID),
		Icon:      "user-x",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

// Deferred delivery

func (s *NotificationServiceImpl) deferNotification(ctx context.Context, userID string, input *NotificationInput, prefs *models.NotificationPreferences) (*models.Notification, error) {
	scheduledFor := s.getQuietHoursEnd(prefs)

	metadataJSON, _ := json.Marshal(input.Metadata)
	var metadata models.JSONMap
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &metadata)
	}
	if metadata == nil {
		metadata = make(models.JSONMap)
	}

	notification := &models.Notification{
		UserID:       userID,
		Type:         input.Type,
		Title:        input.Title,
		Content:      input.Content,
		Metadata:     metadata,
		ActionURL:    input.ActionURL,
		Icon:         input.Icon,
		Priority:     input.Priority,
		IsRead:       false,
		ScheduledFor: &scheduledFor,
		CreatedAt:    time.Now(),
	}

	if notification.Priority == "" {
		notification.Priority = "normal"
	}
	if notification.Icon == "" {
		notification.Icon = s.getIconForType(input.Type)
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create deferred notification: %w", err)
	}

	return notification, nil
}

func (s *NotificationServiceImpl) getQuietHoursEnd(prefs *models.NotificationPreferences) time.Time {
	loc := time.Local
	if prefs.QuietTimezone != "" {
		if l, err := time.LoadLocation(prefs.QuietTimezone); err == nil {
			loc = l
		}
	}
	now := time.Now().In(loc)

	endHour := prefs.QuietEndHour
	endDay := now

	if prefs.QuietStartHour > prefs.QuietEndHour {
		// Overnight (e.g. 22:00 - 08:00)
		if now.Hour() >= prefs.QuietStartHour {
			// Currently in first half of quiet period, ends tomorrow
			endDay = now.AddDate(0, 0, 1)
		}
	}

	return time.Date(endDay.Year(), endDay.Month(), endDay.Day(), endHour, 0, 0, 0, loc)
}

func (s *NotificationServiceImpl) StartDeferredDeliveryRoutine(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		// Run once immediately on start
		s.deliverDeferredNotifications(context.Background())
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.deliverDeferredNotifications(context.Background())
			}
		}
	}()
}

func (s *NotificationServiceImpl) deliverDeferredNotifications(ctx context.Context) {
	notifications, err := s.notificationRepo.ListDeferred(ctx)
	if err != nil {
		return
	}

	for _, n := range notifications {
		prefs, err := s.notificationRepo.GetPreferences(ctx, n.UserID)
		if err != nil {
			continue
		}

		// Re-check quiet hours in case they've changed
		if s.isInQuietHours(prefs) {
			newSchedule := s.getQuietHoursEnd(prefs)
			n.ScheduledFor = &newSchedule
			s.notificationRepo.Update(ctx, n)
			continue
		}

		now := time.Now()
		n.DeliveredAt = &now
		n.ScheduledFor = nil

		if err := s.notificationRepo.Update(ctx, n); err != nil {
			continue
		}

		if prefs.InAppEnabled {
			s.websocketHub.SendToUser(n.UserID, models.WebSocketMessage{
				Type:      "notification",
				Data:      n,
				Timestamp: time.Now(),
			})
		}

		if prefs.EmailEnabled && n.Priority == "high" {
			go s.sendEmailNotification(n, n.UserID)
		}
	}
}

// Helper methods

func (s *NotificationServiceImpl) isNotificationTypeEnabled(prefs *models.NotificationPreferences, notifType string) bool {
	switch notifType {
	case "application":
		return prefs.ApplicationUpdates
	case "interview":
		return prefs.InterviewReminders
	case "alert":
		return prefs.JobAlerts
	case "message":
		return prefs.Messages
	case "system":
		return prefs.SystemAlerts
	case "marketing":
		return prefs.Marketing
	default:
		return true
	}
}

func (s *NotificationServiceImpl) isInQuietHours(prefs *models.NotificationPreferences) bool {
	if !prefs.QuietHoursEnabled {
		return false
	}

	loc := time.Local
	if prefs.QuietTimezone != "" {
		if l, err := time.LoadLocation(prefs.QuietTimezone); err == nil {
			loc = l
		}
	}

	now := time.Now().In(loc)
	currentHour := now.Hour()

	if prefs.QuietStartHour > prefs.QuietEndHour {
		return currentHour >= prefs.QuietStartHour || currentHour < prefs.QuietEndHour
	}

	return currentHour >= prefs.QuietStartHour && currentHour < prefs.QuietEndHour
}

func (s *NotificationServiceImpl) getIconForType(notifType string) string {
	icons := map[string]string{
		"application": "briefcase",
		"interview":   "calendar",
		"alert":       "bell",
		"message":     "message",
		"system":      "settings",
		"marketing":   "megaphone",
	}

	if icon, ok := icons[notifType]; ok {
		return icon
	}
	return "bell"
}

func (s *NotificationServiceImpl) sendEmailNotification(notification *models.Notification, userID string) {
	user, err := s.userRepo.GetByID(context.Background(), userID)
	if err != nil || user == nil {
		return
	}

	actionURL := notification.ActionURL
	if actionURL != "" && !strings.HasPrefix(actionURL, "http") {
		actionURL = s.cfg.AppURL + actionURL
	}

	tmpl := template.Must(template.New("email").Parse(`
		<div style="font-family: sans-serif; padding: 20px; max-width: 600px; margin: 0 auto;">
			<div style="background: #f9fafb; border-radius: 12px; padding: 24px; margin: 20px 0; border: 1px solid #e5e7eb;">
				<h2 style="margin: 0 0 12px; font-size: 20px; color: #111827;">{{.Title}}</h2>
				<p style="margin: 0; font-size: 16px; line-height: 1.6; color: #4b5563;">{{.Content}}</p>
			</div>
			{{if .ActionURL}}
			<div style="text-align: center; margin: 24px 0;">
				<a href="{{.ActionURL}}" style="display: inline-block; padding: 14px 28px; background: #667eea; color: #ffffff; text-decoration: none; border-radius: 8px; font-weight: 600;">View Details</a>
			</div>
			{{end}}
			<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 20px 0;">
			<p style="color: #6b7280; font-size: 12px;">You received this because notifications are enabled on your account.</p>
		</div>
	`))

	var htmlBuf bytes.Buffer
	tmpl.Execute(&htmlBuf, map[string]string{
		"Title":     notification.Title,
		"Content":   notification.Content,
		"ActionURL": actionURL,
	})
	htmlBody := htmlBuf.String()
	textBody := notification.Title + "\n\n" + notification.Content

	s.emailService.SendNotificationEmail(user.Email, notification.Title, htmlBody, textBody, user.Email)
}

// Admin action notification triggers

func (s *NotificationServiceImpl) NotifyJobApproved(ctx context.Context, employerID string, jobID string, jobTitle string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Job Approved",
		Content:   fmt.Sprintf("Your job \"%s\" has been approved and is now live", jobTitle),
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employer/jobs/%s", jobID),
		Icon:      "check-circle",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyJobRejected(ctx context.Context, employerID string, jobID string, jobTitle string, reason string) error {
	content := fmt.Sprintf("Your job \"%s\" has been rejected", jobTitle)
	if reason != "" {
		content += ": " + reason
	}
	input := &NotificationInput{
		Type:      "system",
		Title:     "Job Rejected",
		Content:   content,
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employer/jobs/%s", jobID),
		Icon:      "x-circle",
	}
	_, err := s.SendNotification(ctx, employerID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyContentApproved(ctx context.Context, reporterID string, entityType string, entityID string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Report Resolved",
		Content:   fmt.Sprintf("Your report on %s has been reviewed and resolved", entityType),
		Priority:  "normal",
		ActionURL: "/reports",
		Icon:      "flag",
	}
	_, err := s.SendNotification(ctx, reporterID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyContentRejected(ctx context.Context, reporterID string, entityType string, entityID string, reason string) error {
	content := fmt.Sprintf("Your report on %s was reviewed", entityType)
	if reason != "" {
		content += ": " + reason
	}
	input := &NotificationInput{
		Type:      "system",
		Title:     "Report Update",
		Content:   content,
		Priority:  "normal",
		ActionURL: "/reports",
		Icon:      "flag",
	}
	_, err := s.SendNotification(ctx, reporterID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyCompanyVerified(ctx context.Context, companyID string, companyName string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Company Verified",
		Content:   fmt.Sprintf("Your company \"%s\" has been verified successfully", companyName),
		Priority:  "high",
		ActionURL: fmt.Sprintf("/employer/company"),
		Icon:      "shield-check",
	}
	_, err := s.SendNotification(ctx, companyID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyCompanyRejected(ctx context.Context, companyID string, companyName string, reason string) error {
	content := fmt.Sprintf("Your company \"%s\" verification was rejected", companyName)
	if reason != "" {
		content += ": " + reason
	}
	input := &NotificationInput{
		Type:      "system",
		Title:     "Verification Update",
		Content:   content,
		Priority:  "high",
		ActionURL: "/employer/settings",
		Icon:      "shield-off",
	}
	_, err := s.SendNotification(ctx, companyID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyUserSuspended(ctx context.Context, userID string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Account Suspended",
		Content:   "Your account has been suspended. Please contact support for more information.",
		Priority:  "high",
		ActionURL: "/contact",
		Icon:      "alert-triangle",
	}
	_, err := s.SendNotification(ctx, userID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyUserActivated(ctx context.Context, userID string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "Account Reactivated",
		Content:   "Your account has been reactivated. You can now use all features normally.",
		Priority:  "high",
		ActionURL: "/dashboard",
		Icon:      "check-circle",
	}
	_, err := s.SendNotification(ctx, userID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyAdminNewReport(ctx context.Context, adminID string, entityType string, entityID string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "New Content Report",
		Content:   fmt.Sprintf("A new report has been submitted for %s", entityType),
		Priority:  "high",
		ActionURL: fmt.Sprintf("/admin/reports?entity_type=%s", entityType),
		Icon:      "flag",
	}
	_, err := s.SendNotification(ctx, adminID, input)
	return err
}

func (s *NotificationServiceImpl) NotifyAdminNewVerificationRequest(ctx context.Context, adminID string, companyName string) error {
	input := &NotificationInput{
		Type:      "system",
		Title:     "New Verification Request",
		Content:   fmt.Sprintf("Company \"%s\" has submitted a verification request", companyName),
		Priority:  "high",
		ActionURL: "/admin/companies?status=pending",
		Icon:      "shield",
	}
	_, err := s.SendNotification(ctx, adminID, input)
	return err
}
