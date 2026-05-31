// internal/repository/unsubscribe_repo.go
package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	//"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type UnsubscribeRepository interface {
	CreateToken(ctx context.Context, token *models.UnsubscribeToken) error
	GetToken(ctx context.Context, token string) (*models.UnsubscribeToken, error)
	DeactivateToken(ctx context.Context, token string) error
	GetActiveTokenByEmail(ctx context.Context, email string) (*models.UnsubscribeToken, error)
	CreatePreferences(ctx context.Context, prefs *models.EmailPreferences) error
	GetPreferences(ctx context.Context, userID string) (*models.EmailPreferences, error)
	UpdatePreferences(ctx context.Context, userID string, updates map[string]interface{}) error
	UnsubscribeAll(ctx context.Context, email string) error
	CleanExpiredTokens(ctx context.Context) error
}

type UnsubscribeRepositoryImpl struct {
	db *gorm.DB
}

func NewUnsubscribeRepository(db *gorm.DB) UnsubscribeRepository {
	return &UnsubscribeRepositoryImpl{db: db}
}

func (r *UnsubscribeRepositoryImpl) CreateToken(ctx context.Context, token *models.UnsubscribeToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *UnsubscribeRepositoryImpl) GetToken(ctx context.Context, token string) (*models.UnsubscribeToken, error) {
	var unsubscribeToken models.UnsubscribeToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND is_active = ? AND expires_at > ?", token, true, time.Now()).
		First(&unsubscribeToken).Error
	if err != nil {
		return nil, err
	}
	return &unsubscribeToken, nil
}

func (r *UnsubscribeRepositoryImpl) DeactivateToken(ctx context.Context, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.UnsubscribeToken{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_active":        false,
			"unsubscribed_at":  &now,
			"updated_at":       now,
		}).Error
}

func (r *UnsubscribeRepositoryImpl) GetActiveTokenByEmail(ctx context.Context, email string) (*models.UnsubscribeToken, error) {
	var token models.UnsubscribeToken
	err := r.db.WithContext(ctx).
		Where("email = ? AND is_active = ? AND expires_at > ?", email, true, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *UnsubscribeRepositoryImpl) CreatePreferences(ctx context.Context, prefs *models.EmailPreferences) error {
	return r.db.WithContext(ctx).Create(prefs).Error
}

func (r *UnsubscribeRepositoryImpl) GetPreferences(ctx context.Context, userID string) (*models.EmailPreferences, error) {
	var prefs models.EmailPreferences
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&prefs).Error
	if err != nil {
		return nil, err
	}
	return &prefs, nil
}

func (r *UnsubscribeRepositoryImpl) UpdatePreferences(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&models.EmailPreferences{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *UnsubscribeRepositoryImpl) UnsubscribeAll(ctx context.Context, email string) error {
	// Get user by email to update preferences
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == nil {
		// Update user preferences
		updates := map[string]interface{}{
			"job_alerts":          false,
			"application_updates": false,
			"marketing_emails":    false,
			"newsletter":          false,
			"updated_at":          time.Now(),
		}
		r.db.WithContext(ctx).
			Model(&models.EmailPreferences{}).
			Where("user_id = ?", user.ID).
			Updates(updates)
	}
	
	// Deactivate all active tokens for this email
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.UnsubscribeToken{}).
		Where("email = ? AND is_active = ?", email, true).
		Updates(map[string]interface{}{
			"is_active":        false,
			"unsubscribed_at":  &now,
		}).Error
}

func (r *UnsubscribeRepositoryImpl) CleanExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.UnsubscribeToken{}).Error
}