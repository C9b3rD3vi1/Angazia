package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
	"github.com/google/uuid"
)

type TestimonialService interface {
	List(ctx context.Context, params *models.ListTestimonialsParams) (*models.TestimonialListResponse, error)
	ListMyTestimonials(ctx context.Context, userID string, page, limit int) (*models.TestimonialListResponse, error)
	ListApproved(ctx context.Context, page, limit int) ([]*models.Testimonial, error)
	GetByID(ctx context.Context, id string) (*models.Testimonial, error)
	Create(ctx context.Context, userID string, req *models.CreateTestimonialRequest, role string) (*models.Testimonial, error)
	Update(ctx context.Context, id, userID string, req *models.UpdateTestimonialRequest) (*models.Testimonial, error)
	Delete(ctx context.Context, id string) error
	DeleteOwn(ctx context.Context, id, userID string) error
	Approve(ctx context.Context, id string) error
	Reject(ctx context.Context, id string) error
	ToggleFeatured(ctx context.Context, id string) error
}

type TestimonialServiceImpl struct {
	repo            repository.TestimonialRepository
	userRepo        repository.UserRepository
	adminRepo       repository.AdminRepository
	notificationSvc NotificationService
	emailSvc        EmailService
}

func NewTestimonialService(repo repository.TestimonialRepository, userRepo repository.UserRepository, adminRepo repository.AdminRepository) TestimonialService {
	return &TestimonialServiceImpl{
		repo:      repo,
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

func (s *TestimonialServiceImpl) SetNotificationService(ns NotificationService) {
	s.notificationSvc = ns
}

func (s *TestimonialServiceImpl) SetEmailService(es EmailService) {
	s.emailSvc = es
}

func (s *TestimonialServiceImpl) List(ctx context.Context, params *models.ListTestimonialsParams) (*models.TestimonialListResponse, error) {
	return s.repo.List(ctx, params)
}

func (s *TestimonialServiceImpl) ListMyTestimonials(ctx context.Context, userID string, page, limit int) (*models.TestimonialListResponse, error) {
	return s.repo.ListByUser(ctx, userID, page, limit)
}

func (s *TestimonialServiceImpl) ListApproved(ctx context.Context, page, limit int) ([]*models.Testimonial, error) {
	return s.repo.ListApproved(ctx, page, limit)
}

func (s *TestimonialServiceImpl) GetByID(ctx context.Context, id string) (*models.Testimonial, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TestimonialServiceImpl) Create(ctx context.Context, userID string, req *models.CreateTestimonialRequest, role string) (*models.Testimonial, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	userName := user.Email

	t := &models.Testimonial{
		ID:          uuid.New().String(),
		UserID:      userID,
		UserName:    userName,
		UserTitle:   req.UserTitle,
		CompanyName: req.CompanyName,
		Content:     req.Content,
		Rating:      req.Rating,
		Role:        role,
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}

	if s.notificationSvc != nil {
		adminIDs, err := s.adminRepo.GetAdminUserIDs(ctx)
		if err == nil {
			for _, adminID := range adminIDs {
				s.notificationSvc.SendNotification(ctx, adminID, &NotificationInput{
					Type:      "system",
					Title:     "New Testimonial Submitted",
					Content:   fmt.Sprintf("A new testimonial has been submitted by %s and is pending review", userName),
					Priority:  "high",
					ActionURL: "/admin/testimonials",
					Icon:      "message-square",
				})
			}
		}
	}

	return t, nil
}

func (s *TestimonialServiceImpl) Update(ctx context.Context, id, userID string, req *models.UpdateTestimonialRequest) (*models.Testimonial, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("testimonial not found")
	}
	if t.UserID != userID {
		return nil, errors.New("not authorized to update this testimonial")
	}

	if req.Content != nil {
		t.Content = *req.Content
	}
	if req.UserTitle != nil {
		t.UserTitle = *req.UserTitle
	}
	if req.CompanyName != nil {
		t.CompanyName = *req.CompanyName
	}
	if req.Rating != nil {
		t.Rating = *req.Rating
	}

	t.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TestimonialServiceImpl) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *TestimonialServiceImpl) DeleteOwn(ctx context.Context, id, userID string) error {
	return s.repo.DeleteByUser(ctx, id, userID)
}

func (s *TestimonialServiceImpl) Approve(ctx context.Context, id string) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Approve(ctx, id); err != nil {
		return err
	}

	if s.notificationSvc != nil {
		s.notificationSvc.SendNotification(ctx, t.UserID, &NotificationInput{
			Type:      "system",
			Title:     "Testimonial Approved",
			Content:   "Your testimonial has been approved and is now visible on Angazia",
			Priority:  "high",
			ActionURL: "/employee/testimonials",
			Icon:      "check-circle",
		})
	}

	if s.emailSvc != nil {
		user, err := s.userRepo.GetByID(ctx, t.UserID)
		if err == nil && user != nil {
			go s.sendTestimonialApprovalEmail(user.Email, user.Email)
		}
	}

	return nil
}

func (s *TestimonialServiceImpl) sendTestimonialApprovalEmail(to, email string) {
	data := map[string]interface{}{
		"AppName":          "Angazia",
		"AppURL":           "https://angazia.co.ke",
		"HeaderTitle":      "Testimonial Approved!",
		"HeaderColor":      "linear-gradient(135deg, #10b981, #059669)",
		"ButtonColor":      "#10b981",
		"ButtonHoverColor": "#059669",
		"LinkColor":        "#10b981",
		"Year":             time.Now().Year(),
		"Email":            email,
	}
	s.emailSvc.SendNotificationEmail(
		to, "Your testimonial has been approved - Angazia",
		s.renderTestimonialEmailHTML("testimonial_approved", data),
		"Your testimonial has been approved and is now visible on Angazia.",
		email,
	)
}

func (s *TestimonialServiceImpl) Reject(ctx context.Context, id string) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Reject(ctx, id); err != nil {
		return err
	}

	if s.notificationSvc != nil {
		s.notificationSvc.SendNotification(ctx, t.UserID, &NotificationInput{
			Type:      "system",
			Title:     "Testimonial Update",
			Content:   "Your testimonial was not approved after review. Please contact support if you have questions.",
			Priority:  "high",
			ActionURL: "/employee/testimonials",
			Icon:      "x-circle",
		})
	}

	if s.emailSvc != nil {
		user, err := s.userRepo.GetByID(ctx, t.UserID)
		if err == nil && user != nil {
			go s.sendTestimonialRejectedEmail(user.Email, user.Email)
		}
	}

	return nil
}

func (s *TestimonialServiceImpl) sendTestimonialRejectedEmail(to, email string) {
	data := map[string]interface{}{
		"AppName":          "Angazia",
		"AppURL":           "https://angazia.co.ke",
		"HeaderTitle":      "Testimonial Update",
		"HeaderColor":      "linear-gradient(135deg, #f59e0b, #d97706)",
		"ButtonColor":      "#f59e0b",
		"ButtonHoverColor": "#d97706",
		"LinkColor":        "#f59e0b",
		"Year":             time.Now().Year(),
		"Email":            email,
	}
	s.emailSvc.SendNotificationEmail(
		to, "Update on your testimonial - Angazia",
		s.renderTestimonialEmailHTML("testimonial_rejected", data),
		"Your testimonial was not approved after review. Please contact support if you have questions.",
		email,
	)
}

func (s *TestimonialServiceImpl) renderTestimonialEmailHTML(tmplName string, data map[string]interface{}) string {
	raw := "<p>Thank you for using Angazia.</p>"
	return raw
}

func (s *TestimonialServiceImpl) ToggleFeatured(ctx context.Context, id string) error {
	return s.repo.ToggleFeatured(ctx, id)
}
