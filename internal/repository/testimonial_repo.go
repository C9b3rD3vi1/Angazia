package repository

import (
	"context"
	"math"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"gorm.io/gorm"
)

type TestimonialRepository interface {
	List(ctx context.Context, params *models.ListTestimonialsParams) (*models.TestimonialListResponse, error)
	ListByUser(ctx context.Context, userID string, page, limit int) (*models.TestimonialListResponse, error)
	ListApproved(ctx context.Context, page, limit int) ([]*models.Testimonial, error)
	GetByID(ctx context.Context, id string) (*models.Testimonial, error)
	Create(ctx context.Context, t *models.Testimonial) error
	Update(ctx context.Context, t *models.Testimonial) error
	Delete(ctx context.Context, id string) error
	Approve(ctx context.Context, id string) error
	Reject(ctx context.Context, id string) error
	ToggleFeatured(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, id, userID string) error
	CountByUser(ctx context.Context, userID string) (int64, error)
}

type TestimonialRepositoryImpl struct {
	db *gorm.DB
}

func NewTestimonialRepository(db *gorm.DB) TestimonialRepository {
	return &TestimonialRepositoryImpl{db: db}
}

func (r *TestimonialRepositoryImpl) List(ctx context.Context, params *models.ListTestimonialsParams) (*models.TestimonialListResponse, error) {
	query := r.db.WithContext(ctx).Model(&models.Testimonial{})

	if params.Status == "approved" {
		query = query.Where("is_approved = ?", true)
	} else if params.Status == "pending" {
		query = query.Where("is_approved = ?", false)
	}

	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
	}

	if params.IsFeatured != nil {
		query = query.Where("is_featured = ?", *params.IsFeatured)
	}

	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("user_name ILIKE ? OR content ILIKE ? OR company_name ILIKE ?", search, search, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	page := params.Page
	limit := params.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var testimonials []*models.Testimonial
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&testimonials).Error; err != nil {
		return nil, err
	}

	return &models.TestimonialListResponse{
		Testimonials: testimonials,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}, nil
}

func (r *TestimonialRepositoryImpl) ListByUser(ctx context.Context, userID string, page, limit int) (*models.TestimonialListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Testimonial{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, err
	}

	var testimonials []*models.Testimonial
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&testimonials).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &models.TestimonialListResponse{
		Testimonials: testimonials,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}, nil
}

func (r *TestimonialRepositoryImpl) ListApproved(ctx context.Context, page, limit int) ([]*models.Testimonial, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var testimonials []*models.Testimonial
	if err := r.db.WithContext(ctx).
		Where("is_approved = ?", true).
		Order("is_featured DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&testimonials).Error; err != nil {
		return nil, err
	}
	return testimonials, nil
}

func (r *TestimonialRepositoryImpl) GetByID(ctx context.Context, id string) (*models.Testimonial, error) {
	var t models.Testimonial
	if err := r.db.WithContext(ctx).Preload("User").First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TestimonialRepositoryImpl) Create(ctx context.Context, t *models.Testimonial) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TestimonialRepositoryImpl) Update(ctx context.Context, t *models.Testimonial) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *TestimonialRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Testimonial{}, "id = ?", id).Error
}

func (r *TestimonialRepositoryImpl) Approve(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.Testimonial{}).Where("id = ?", id).Update("is_approved", true).Error
}

func (r *TestimonialRepositoryImpl) Reject(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Testimonial{}, "id = ?", id).Error
}

func (r *TestimonialRepositoryImpl) ToggleFeatured(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Exec("UPDATE testimonials SET is_featured = NOT is_featured WHERE id = ?", id).Error
}

func (r *TestimonialRepositoryImpl) DeleteByUser(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.Testimonial{}).Error
}

func (r *TestimonialRepositoryImpl) CountByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Testimonial{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
