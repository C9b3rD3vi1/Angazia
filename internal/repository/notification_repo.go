package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type NotificationRepository interface {
	// Notification CRUD
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id, userID string) (*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
	Delete(ctx context.Context, id, userID string) error
	DeleteAll(ctx context.Context, userID string) error
	
	// List notifications
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*models.Notification, int64, error)
	ListUnread(ctx context.Context, userID string, limit int) ([]*models.Notification, error)
	ListByType(ctx context.Context, userID, notifType string, page, limit int) ([]*models.Notification, int64, error)
	ListByPriority(ctx context.Context, userID, priority string, page, limit int) ([]*models.Notification, int64, error)
	
	// Status updates
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkMultipleAsRead(ctx context.Context, ids []string, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Archive(ctx context.Context, id, userID string) error
	
	// Counts
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	GetUnreadCountByType(ctx context.Context, userID string) (map[string]int, error)
	
	// Preferences
	GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error)
	CreatePreferences(ctx context.Context, prefs *models.NotificationPreferences) error
	UpdatePreferences(ctx context.Context, userID string, updates map[string]interface{}) error
	UpsertPreferences(ctx context.Context, prefs *models.NotificationPreferences) error
	
	// Cleanup
	DeleteOldNotifications(ctx context.Context, days int) error
}

type NotificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &NotificationRepositoryImpl{db: db}
}

func (r *NotificationRepositoryImpl) Create(ctx context.Context, notification *models.Notification) error {
	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *NotificationRepositoryImpl) GetByID(ctx context.Context, id, userID string) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepositoryImpl) Update(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *NotificationRepositoryImpl) Delete(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Notification{}).Error
}

func (r *NotificationRepositoryImpl) DeleteAll(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.Notification{}).Error
}

func (r *NotificationRepositoryImpl) ListByUser(ctx context.Context, userID string, page, limit int) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND is_archived = ?", userID, false)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error
	
	return notifications, total, err
}

func (r *NotificationRepositoryImpl) ListUnread(ctx context.Context, userID string, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ? AND is_archived = ?", userID, false, false).
		Order("priority DESC, created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepositoryImpl) ListByType(ctx context.Context, userID, notifType string, page, limit int) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND type = ? AND is_archived = ?", userID, notifType, false)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error
	
	return notifications, total, err
}

func (r *NotificationRepositoryImpl) ListByPriority(ctx context.Context, userID, priority string, page, limit int) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND priority = ? AND is_archived = ?", userID, priority, false)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error
	
	return notifications, total, err
}

func (r *NotificationRepositoryImpl) MarkAsRead(ctx context.Context, id, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *NotificationRepositoryImpl) MarkMultipleAsRead(ctx context.Context, ids []string, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *NotificationRepositoryImpl) MarkAllAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *NotificationRepositoryImpl) Archive(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_archived", true).Error
}

func (r *NotificationRepositoryImpl) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ? AND is_archived = ?", userID, false, false).
		Count(&count).Error
	return int(count), err
}

func (r *NotificationRepositoryImpl) GetUnreadCountByType(ctx context.Context, userID string) (map[string]int, error) {
	var results []struct {
		Type  string
		Count int
	}
	
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ? AND is_read = ? AND is_archived = ?", userID, false, false).
		Group("type").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	counts := make(map[string]int)
	for _, r := range results {
		counts[r.Type] = r.Count
	}
	
	return counts, nil
}

func (r *NotificationRepositoryImpl) GetPreferences(ctx context.Context, userID string) (*models.NotificationPreferences, error) {
	var prefs models.NotificationPreferences
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&prefs).Error
	if err != nil {
		return nil, err
	}
	return &prefs, nil
}

func (r *NotificationRepositoryImpl) CreatePreferences(ctx context.Context, prefs *models.NotificationPreferences) error {
	prefs.ID = uuid.New().String()
	prefs.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(prefs).Error
}

func (r *NotificationRepositoryImpl) UpdatePreferences(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&models.NotificationPreferences{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *NotificationRepositoryImpl) UpsertPreferences(ctx context.Context, prefs *models.NotificationPreferences) error {
	var existing models.NotificationPreferences
	err := r.db.WithContext(ctx).Where("user_id = ?", prefs.UserID).First(&existing).Error
	
	if err == nil {
		prefs.ID = existing.ID
		prefs.CreatedAt = existing.CreatedAt
		return r.UpdatePreferences(ctx, prefs.UserID, map[string]interface{}{
			"push_enabled":        prefs.PushEnabled,
			"push_sound":          prefs.PushSound,
			"email_enabled":       prefs.EmailEnabled,
			"email_digest":        prefs.EmailDigest,
			"in_app_enabled":      prefs.InAppEnabled,
			"application_updates": prefs.ApplicationUpdates,
			"job_alerts":          prefs.JobAlerts,
			"interview_reminders": prefs.InterviewReminders,
			"messages":            prefs.Messages,
			"system_alerts":       prefs.SystemAlerts,
			"marketing":           prefs.Marketing,
			"quiet_hours_enabled": prefs.QuietHoursEnabled,
			"quiet_start_hour":    prefs.QuietStartHour,
			"quiet_end_hour":      prefs.QuietEndHour,
			"quiet_timezone":      prefs.QuietTimezone,
		})
	}
	
	return r.CreatePreferences(ctx, prefs)
}

func (r *NotificationRepositoryImpl) DeleteOldNotifications(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.WithContext(ctx).
		Where("created_at < ? AND is_read = ?", cutoffDate, true).
		Delete(&models.Notification{}).Error
}