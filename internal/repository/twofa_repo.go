package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type TwoFARepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.TwoFASecret, error)
	Upsert(ctx context.Context, secret *models.TwoFASecret) error
	Update(ctx context.Context, secret *models.TwoFASecret) error
	Delete(ctx context.Context, userID string) error
	LogEvent(ctx context.Context, log *models.TwoFAAuditLog) error
}

type TwoFARepositoryImpl struct {
	db *gorm.DB
}

func NewTwoFARepository(db *gorm.DB) TwoFARepository {
	return &TwoFARepositoryImpl{db: db}
}

func (r *TwoFARepositoryImpl) GetByUserID(ctx context.Context, userID string) (*models.TwoFASecret, error) {
	var secret models.TwoFASecret
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&secret).Error
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *TwoFARepositoryImpl) Upsert(ctx context.Context, secret *models.TwoFASecret) error {
	var existing models.TwoFASecret
	err := r.db.WithContext(ctx).Where("user_id = ?", secret.UserID).First(&existing).Error
	
	if err == nil {
		secret.ID = existing.ID
		secret.CreatedAt = existing.CreatedAt
		return r.Update(ctx, secret)
	}
	
	secret.ID = uuid.New().String()
	secret.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(secret).Error
}

func (r *TwoFARepositoryImpl) Update(ctx context.Context, secret *models.TwoFASecret) error {
	secret.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(secret).Error
}

func (r *TwoFARepositoryImpl) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.TwoFASecret{}).Error
}

func (r *TwoFARepositoryImpl) LogEvent(ctx context.Context, log *models.TwoFAAuditLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(log).Error
}