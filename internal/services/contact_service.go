package services

import (
	"context"
	"fmt"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type ContactService interface {
	Submit(ctx context.Context, name, email, subject, message string) error
	List(ctx context.Context, page, limit int, search string) (*models.ContactSubmissionListResponse, error)
	GetByID(ctx context.Context, id string) (*models.ContactSubmission, error)
	MarkAsRead(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	GetUnreadCount(ctx context.Context) (int64, error)
}

type ContactServiceImpl struct {
	repo            repository.ContactRepository
	adminRepo       repository.AdminRepository
	userRepo        repository.UserRepository
	notificationSvc NotificationService
	emailSvc        EmailService
}

func NewContactService(
	repo repository.ContactRepository,
	adminRepo repository.AdminRepository,
	userRepo repository.UserRepository,
) *ContactServiceImpl {
	return &ContactServiceImpl{
		repo:      repo,
		adminRepo: adminRepo,
		userRepo:  userRepo,
	}
}

func (s *ContactServiceImpl) SetNotificationService(ns NotificationService) {
	s.notificationSvc = ns
}

func (s *ContactServiceImpl) SetEmailService(es EmailService) {
	s.emailSvc = es
}

func (s *ContactServiceImpl) Submit(ctx context.Context, name, email, subject, message string) error {
	sub := &models.ContactSubmission{
		Name:    name,
		Email:   email,
		Subject: subject,
		Message: message,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return fmt.Errorf("failed to save contact submission: %w", err)
	}

	// Notify all admins
	if s.notificationSvc != nil {
		adminIDs, err := s.adminRepo.GetAdminUserIDs(ctx)
		if err == nil && len(adminIDs) > 0 {
			for _, adminID := range adminIDs {
				input := &NotificationInput{
					Type:      "system",
					Title:     "New Contact Message",
					Content:   fmt.Sprintf("%s (%s) sent a message: %s", name, email, subject),
					Priority:  "high",
					ActionURL: "/admin/contacts",
					Icon:      "mail",
				}
				s.notificationSvc.SendNotification(ctx, adminID, input)
			}
		}
	}

	// Send auto-reply email
	if s.emailSvc != nil {
		s.emailSvc.SendNotificationEmail(
			email,
			"Thank you for contacting "+s.getAppName(),
			s.buildAutoReplyHTML(name),
			fmt.Sprintf("Thank you for reaching out, %s. We have received your message and will get back to you within 24 hours.", name),
			email,
		)
	}

	return nil
}

func (s *ContactServiceImpl) getAppName() string {
	return "Angazia"
}

func (s *ContactServiceImpl) buildAutoReplyHTML(name string) string {
	return fmt.Sprintf(`<div style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:20px">
		<div style="background:linear-gradient(135deg,#667eea,#764ba2);color:white;padding:32px;text-align:center;border-radius:12px 12px 0 0">
			<h1 style="margin:0;font-size:24px">Thank You, %s!</h1>
		</div>
		<div style="background:#fff;padding:32px;border:1px solid #e5e7eb;border-top:none;border-radius:0 0 12px 12px">
			<p style="font-size:16px;line-height:1.6;color:#374151">We have received your message and our team will review it shortly.</p>
			<p style="font-size:16px;line-height:1.6;color:#374151">We typically respond within <strong>24 hours</strong> during business days.</p>
			<div style="background:#f9fafb;padding:20px;border-radius:8px;margin:20px 0;border:1px solid #e5e7eb">
				<p style="margin:0;font-size:14px;color:#6b7280">In the meantime, feel free to explore our platform and check out available opportunities.</p>
			</div>
			<hr style="border:none;border-top:1px solid #e5e7eb;margin:24px 0">
			<p style="font-size:14px;color:#6b7280">Best regards,<br>The Angazia Team</p>
		</div>
	</div>`, name)
}

func (s *ContactServiceImpl) List(ctx context.Context, page, limit int, search string) (*models.ContactSubmissionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, page, limit, search)
}

func (s *ContactServiceImpl) GetByID(ctx context.Context, id string) (*models.ContactSubmission, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ContactServiceImpl) MarkAsRead(ctx context.Context, id string) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *ContactServiceImpl) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ContactServiceImpl) GetUnreadCount(ctx context.Context) (int64, error) {
	return s.repo.GetUnreadCount(ctx)
}
