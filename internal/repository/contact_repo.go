package repository

import (
	"context"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContactRepository interface {
	Create(ctx context.Context, sub *models.ContactSubmission) error
	List(ctx context.Context, page, limit int, search string) (*models.ContactSubmissionListResponse, error)
	GetByID(ctx context.Context, id string) (*models.ContactSubmission, error)
	MarkAsRead(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	GetUnreadCount(ctx context.Context) (int64, error)
}

type ContactRepositoryImpl struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) ContactRepository {
	return &ContactRepositoryImpl{db: db}
}

func (r *ContactRepositoryImpl) Create(ctx context.Context, sub *models.ContactSubmission) error {
	sub.ID = uuid.New().String()
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *ContactRepositoryImpl) List(ctx context.Context, page, limit int, search string) (*models.ContactSubmissionListResponse, error) {
	query := r.db.WithContext(ctx).Model(&models.ContactSubmission{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR subject ILIKE ? OR message ILIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	var submissions []*models.ContactSubmission
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&submissions).Error; err != nil {
		return nil, err
	}

	var unreadCount int64
	r.db.WithContext(ctx).Model(&models.ContactSubmission{}).Where("is_read = ?", false).Count(&unreadCount)

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &models.ContactSubmissionListResponse{
		Submissions: submissions,
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
		UnreadCount: unreadCount,
	}, nil
}

func (r *ContactRepositoryImpl) GetByID(ctx context.Context, id string) (*models.ContactSubmission, error) {
	var sub models.ContactSubmission
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *ContactRepositoryImpl) MarkAsRead(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.ContactSubmission{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *ContactRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ContactSubmission{}).Error
}

func (r *ContactRepositoryImpl) GetUnreadCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ContactSubmission{}).Where("is_read = ?", false).Count(&count).Error
	return count, err
}
