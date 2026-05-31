package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type AlertRepository interface {
	// Saved search operations
	CreateSavedSearch(ctx context.Context, search *models.SavedSearch) error
	GetSavedSearch(ctx context.Context, id string) (*models.SavedSearch, error)
	GetSavedSearchByUser(ctx context.Context, id, userID string) (*models.SavedSearch, error)
	UpdateSavedSearch(ctx context.Context, search *models.SavedSearch) error
	DeleteSavedSearch(ctx context.Context, id, userID string) error
	ListSavedSearchesByUser(ctx context.Context, userID string, page, limit int) ([]*models.SavedSearch, int64, error)
	ListActiveSavedSearches(ctx context.Context) ([]*models.SavedSearch, error)
	UpdateLastSent(ctx context.Context, id string, sentAt time.Time) error
	
	// Alert history operations
	CreateAlertHistory(ctx context.Context, history *models.AlertHistory) error
	GetAlertHistoryBySearch(ctx context.Context, searchID string, limit int) ([]*models.AlertHistory, error)
	GetRecentAlerts(ctx context.Context, userID string, days int) ([]*models.AlertHistory, error)
	
	// Alert settings operations
	GetAlertSettings(ctx context.Context, userID string) (*models.AlertSettings, error)
	CreateAlertSettings(ctx context.Context, settings *models.AlertSettings) error
	UpdateAlertSettings(ctx context.Context, userID string, updates map[string]interface{}) error
	UpsertAlertSettings(ctx context.Context, settings *models.AlertSettings) error
}

type AlertRepositoryImpl struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &AlertRepositoryImpl{db: db}
}

func (r *AlertRepositoryImpl) CreateSavedSearch(ctx context.Context, search *models.SavedSearch) error {
	search.ID = uuid.New().String()
	search.CreatedAt = time.Now()
	search.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(search).Error
}

func (r *AlertRepositoryImpl) GetSavedSearch(ctx context.Context, id string) (*models.SavedSearch, error) {
	var search models.SavedSearch
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&search).Error
	if err != nil {
		return nil, err
	}
	return &search, nil
}

func (r *AlertRepositoryImpl) GetSavedSearchByUser(ctx context.Context, id, userID string) (*models.SavedSearch, error) {
	var search models.SavedSearch
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&search).Error
	if err != nil {
		return nil, err
	}
	return &search, nil
}

func (r *AlertRepositoryImpl) UpdateSavedSearch(ctx context.Context, search *models.SavedSearch) error {
	search.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(search).Error
}

func (r *AlertRepositoryImpl) DeleteSavedSearch(ctx context.Context, id, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.SavedSearch{}).Error
}

func (r *AlertRepositoryImpl) ListSavedSearchesByUser(ctx context.Context, userID string, page, limit int) ([]*models.SavedSearch, int64, error) {
	var searches []*models.SavedSearch
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.SavedSearch{}).Where("user_id = ?", userID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&searches).Error
	
	return searches, total, err
}

func (r *AlertRepositoryImpl) ListActiveSavedSearches(ctx context.Context) ([]*models.SavedSearch, error) {
	var searches []*models.SavedSearch
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("User").
		Find(&searches).Error
	return searches, err
}

func (r *AlertRepositoryImpl) UpdateLastSent(ctx context.Context, id string, sentAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.SavedSearch{}).
		Where("id = ?", id).
		Update("last_sent_at", sentAt).Error
}

func (r *AlertRepositoryImpl) CreateAlertHistory(ctx context.Context, history *models.AlertHistory) error {
	history.ID = uuid.New().String()
	history.SentAt = time.Now()
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *AlertRepositoryImpl) GetAlertHistoryBySearch(ctx context.Context, searchID string, limit int) ([]*models.AlertHistory, error) {
	var history []*models.AlertHistory
	err := r.db.WithContext(ctx).
		Where("saved_search_id = ?", searchID).
		Order("sent_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

func (r *AlertRepositoryImpl) GetRecentAlerts(ctx context.Context, userID string, days int) ([]*models.AlertHistory, error) {
	since := time.Now().AddDate(0, 0, -days)
	var history []*models.AlertHistory
	err := r.db.WithContext(ctx).
		Joins("JOIN saved_searches ON saved_searches.id = alert_history.saved_search_id").
		Where("saved_searches.user_id = ? AND alert_history.sent_at >= ?", userID, since).
		Order("alert_history.sent_at DESC").
		Find(&history).Error
	return history, err
}

func (r *AlertRepositoryImpl) GetAlertSettings(ctx context.Context, userID string) (*models.AlertSettings, error) {
	var settings models.AlertSettings
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *AlertRepositoryImpl) CreateAlertSettings(ctx context.Context, settings *models.AlertSettings) error {
	settings.ID = uuid.New().String()
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(settings).Error
}

func (r *AlertRepositoryImpl) UpdateAlertSettings(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&models.AlertSettings{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

func (r *AlertRepositoryImpl) UpsertAlertSettings(ctx context.Context, settings *models.AlertSettings) error {
	var existing models.AlertSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", settings.UserID).First(&existing).Error
	
	if err == nil {
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
		settings.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Save(settings).Error
	}
	
	settings.ID = uuid.New().String()
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(settings).Error
}