package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type SubscriptionRepository interface {
	// Subscription CRUD
	CreateSubscription(ctx context.Context, sub *models.Subscription) error
	GetSubscription(ctx context.Context, id string) (*models.Subscription, error)
	GetSubscriptionByUser(ctx context.Context, userID string) (*models.Subscription, error)
	GetActiveSubscriptionByUser(ctx context.Context, userID string) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	CancelSubscription(ctx context.Context, id string, cancelledAt time.Time) error
	DeleteSubscription(ctx context.Context, id string) error
	ListSubscriptions(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Subscription, int64, error)
	
	// Expiry handling
	GetExpiredSubscriptions(ctx context.Context) ([]*models.Subscription, error)
	GetExpiringSoonSubscriptions(ctx context.Context, days int) ([]*models.Subscription, error)
	
	// Subscription Plans
	CreatePlan(ctx context.Context, plan *models.SubscriptionPlan) error
	GetPlan(ctx context.Context, id string) (*models.SubscriptionPlan, error)
	GetPlanByPlanID(ctx context.Context, planID string) (*models.SubscriptionPlan, error)
	GetAllPlans(ctx context.Context, includeInactive bool) ([]*models.SubscriptionPlan, error)
	UpdatePlan(ctx context.Context, plan *models.SubscriptionPlan) error
	DeletePlan(ctx context.Context, id string) error
	
	// Plan Features
	AddPlanFeature(ctx context.Context, feature *models.SubscriptionPlanFeature) error
	GetPlanFeatures(ctx context.Context, planID string) ([]*models.SubscriptionPlanFeature, error)
	UpdatePlanFeature(ctx context.Context, id string, isEnabled bool, featureValue string) error
	DeletePlanFeature(ctx context.Context, id string) error
	
	// Subscription History
	AddHistory(ctx context.Context, history *models.SubscriptionHistory) error
	GetSubscriptionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]*models.SubscriptionHistory, int64, error)
	
	// Usage tracking
	GetOrCreateUsage(ctx context.Context, subscriptionID, userID, metricKey string, limit int, periodStart, periodEnd time.Time) (*models.SubscriptionUsage, error)
	IncrementUsage(ctx context.Context, id string) error
	ResetUsage(ctx context.Context, subscriptionID, metricKey string) error
	GetUsage(ctx context.Context, subscriptionID, metricKey string) (*models.SubscriptionUsage, error)
	GetAllUsage(ctx context.Context, subscriptionID string) ([]*models.SubscriptionUsage, error)
	ResetExpiredUsage(ctx context.Context) error
}

type SubscriptionRepositoryImpl struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &SubscriptionRepositoryImpl{db: db}
}

func (r *SubscriptionRepositoryImpl) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	sub.ID = uuid.New().String()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepositoryImpl) GetSubscription(ctx context.Context, id string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("LastPayment").
		Where("id = ?", id).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepositoryImpl) GetSubscriptionByUser(ctx context.Context, userID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepositoryImpl) GetActiveSubscriptionByUser(ctx context.Context, userID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND end_date > ?", userID, "active", time.Now()).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepositoryImpl) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	sub.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *SubscriptionRepositoryImpl) CancelSubscription(ctx context.Context, id string, cancelledAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "cancelled",
			"cancelled_at": cancelledAt,
			"auto_renew":   false,
			"updated_at":   time.Now(),
		}).Error
}

func (r *SubscriptionRepositoryImpl) DeleteSubscription(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error
}

func (r *SubscriptionRepositoryImpl) ListSubscriptions(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Subscription, int64, error) {
	var subs []*models.Subscription
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Subscription{})
	
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if planID, ok := filters["plan_id"].(string); ok && planID != "" {
		query = query.Where("plan_id = ?", planID)
	}
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&subs).Error
	
	return subs, total, err
}

func (r *SubscriptionRepositoryImpl) GetExpiredSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	var subs []*models.Subscription
	err := r.db.WithContext(ctx).
		Where("status IN ? AND end_date < ?", []string{"active", "trialing"}, time.Now()).
		Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepositoryImpl) GetExpiringSoonSubscriptions(ctx context.Context, days int) ([]*models.Subscription, error) {
	var subs []*models.Subscription
	err := r.db.WithContext(ctx).
		Where("status = ? AND end_date BETWEEN ? AND ?", "active", time.Now(), time.Now().AddDate(0, 0, days)).
		Find(&subs).Error
	return subs, err
}

// CreatePlan creates a new subscription plan
func (r *SubscriptionRepositoryImpl) CreatePlan(ctx context.Context, plan *models.SubscriptionPlan) error {
	plan.ID = uuid.New().String()
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *SubscriptionRepositoryImpl) GetPlan(ctx context.Context, id string) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionRepositoryImpl) GetPlanByPlanID(ctx context.Context, planID string) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	// Remove the is_active filter so we can find inactive plans too
	err := r.db.WithContext(ctx).Where("plan_id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}


func (r *SubscriptionRepositoryImpl) GetAllPlans(ctx context.Context, includeInactive bool) ([]*models.SubscriptionPlan, error) {
	var plans []*models.SubscriptionPlan
	query := r.db.WithContext(ctx)
	
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	
	err := query.Order("sort_order ASC, price ASC").Find(&plans).Error
	return plans, err
}

func (r *SubscriptionRepositoryImpl) UpdatePlan(ctx context.Context, plan *models.SubscriptionPlan) error {
	plan.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *SubscriptionRepositoryImpl) DeletePlan(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.SubscriptionPlan{}, "id = ?", id).Error
}

func (r *SubscriptionRepositoryImpl) AddPlanFeature(ctx context.Context, feature *models.SubscriptionPlanFeature) error {
	feature.ID = uuid.New().String()
	feature.CreatedAt = time.Now()
	feature.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(feature).Error
}

func (r *SubscriptionRepositoryImpl) GetPlanFeatures(ctx context.Context, planID string) ([]*models.SubscriptionPlanFeature, error) {
	var features []*models.SubscriptionPlanFeature
	err := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Find(&features).Error
	return features, err
}

func (r *SubscriptionRepositoryImpl) UpdatePlanFeature(ctx context.Context, id string, isEnabled bool, featureValue string) error {
	updates := map[string]interface{}{
		"is_enabled": isEnabled,
		"updated_at": time.Now(),
	}
	if featureValue != "" {
		updates["feature_value"] = featureValue
	}
	return r.db.WithContext(ctx).
		Model(&models.SubscriptionPlanFeature{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *SubscriptionRepositoryImpl) DeletePlanFeature(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.SubscriptionPlanFeature{}, "id = ?", id).Error
}

func (r *SubscriptionRepositoryImpl) AddHistory(ctx context.Context, history *models.SubscriptionHistory) error {
	history.ID = uuid.New().String()
	history.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *SubscriptionRepositoryImpl) GetSubscriptionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]*models.SubscriptionHistory, int64, error) {
	var history []*models.SubscriptionHistory
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.SubscriptionHistory{}).
		Where("subscription_id = ?", subscriptionID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&history).Error
	
	return history, total, err
}

func (r *SubscriptionRepositoryImpl) GetOrCreateUsage(ctx context.Context, subscriptionID, userID, metricKey string, limit int, periodStart, periodEnd time.Time) (*models.SubscriptionUsage, error) {
	var usage models.SubscriptionUsage
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND metric_key = ? AND period_start = ?", subscriptionID, metricKey, periodStart).
		First(&usage).Error
	
	if err == nil {
		return &usage, nil
	}
	
	usage = models.SubscriptionUsage{
		SubscriptionID: subscriptionID,
		UserID:         userID,
		MetricKey:      metricKey,
		CurrentUsage:   0,
		Limit:          limit,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
	}
	
	if err := r.db.WithContext(ctx).Create(&usage).Error; err != nil {
		return nil, err
	}
	
	return &usage, nil
}

func (r *SubscriptionRepositoryImpl) IncrementUsage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.SubscriptionUsage{}).
		Where("id = ?", id).
		UpdateColumn("current_usage", gorm.Expr("current_usage + 1")).Error
}

func (r *SubscriptionRepositoryImpl) ResetUsage(ctx context.Context, subscriptionID, metricKey string) error {
	return r.db.WithContext(ctx).
		Model(&models.SubscriptionUsage{}).
		Where("subscription_id = ? AND metric_key = ?", subscriptionID, metricKey).
		Update("current_usage", 0).Error
}

func (r *SubscriptionRepositoryImpl) GetUsage(ctx context.Context, subscriptionID, metricKey string) (*models.SubscriptionUsage, error) {
	var usage models.SubscriptionUsage
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND metric_key = ?", subscriptionID, metricKey).
		First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *SubscriptionRepositoryImpl) GetAllUsage(ctx context.Context, subscriptionID string) ([]*models.SubscriptionUsage, error) {
	var usage []*models.SubscriptionUsage
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Find(&usage).Error
	return usage, err
}

func (r *SubscriptionRepositoryImpl) ResetExpiredUsage(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&models.SubscriptionUsage{}).
		Where("period_end < ?", time.Now()).
		Update("current_usage", 0).Error
}