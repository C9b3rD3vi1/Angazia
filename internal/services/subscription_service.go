package services

import (
	"context"
	"fmt"
	"time"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type SubscriptionService interface {
	// Plan management
	GetPlans(ctx context.Context) ([]*models.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, planID string) (*models.SubscriptionPlan, error)
	CreatePlan(ctx context.Context, req *CreatePlanRequest) (*models.SubscriptionPlan, error)
	UpdatePlan(ctx context.Context, planID string, req *UpdatePlanRequest) (*models.SubscriptionPlan, error)
	DeletePlan(ctx context.Context, planID string) error
	
	// Subscription management
	Subscribe(ctx context.Context, userID string, req *SubscribeRequest) (*models.Subscription, error)
	CancelSubscription(ctx context.Context, userID, subscriptionID string, reason string) error
	GetCurrentSubscription(ctx context.Context, userID string) (*models.Subscription, error)
	GetSubscription(ctx context.Context, subscriptionID string) (*models.Subscription, error)
	ListUserSubscriptions(ctx context.Context, userID string, page, limit int) ([]*models.Subscription, int64, error)
	
	// Feature checking
	CanAccessFeature(ctx context.Context, userID, feature string) (bool, error)
	GetRemainingUsage(ctx context.Context, userID, metricKey string) (int, error)
	TrackUsage(ctx context.Context, userID, metricKey string) error
	
	// Billing
	GenerateInvoice(ctx context.Context, subscriptionID, paymentID string) (*models.Invoice, error)
	GetInvoices(ctx context.Context, userID string, page, limit int) ([]*models.Invoice, int64, error)
	
	// Subscription lifecycle
	RenewSubscription(ctx context.Context, subscriptionID string) error
	ExpireSubscriptions(ctx context.Context) error
	UpgradeSubscription(ctx context.Context, userID, subscriptionID, newPlanID string) (*models.Subscription, error)
	DowngradeSubscription(ctx context.Context, userID, subscriptionID, newPlanID string) (*models.Subscription, error)
	
	// Admin
	GetAllSubscriptions(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Subscription, int64, error)
	GetSubscriptionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]*models.SubscriptionHistory, int64, error)
}

type SubscribeRequest struct {
	PlanID     string `json:"plan_id" validate:"required"`
	PaymentID  string `json:"payment_id" validate:"required"`
	TrialDays  int    `json:"trial_days"`
	AutoRenew  bool   `json:"auto_renew"`
}

type CreatePlanRequest struct {
	PlanID         string   `json:"plan_id" validate:"required"`
	Name           string   `json:"name" validate:"required"`
	Description    string   `json:"description"`
	Price          float64  `json:"price" validate:"required,min=0"`
	Currency       string   `json:"currency"`
	Interval       string   `json:"interval" validate:"oneof=month year"`
	IntervalCount  int      `json:"interval_count"`
	TrialDays      int      `json:"trial_days"`
	JobPostLimit   int      `json:"job_post_limit"`
	SortOrder      int      `json:"sort_order"`
	IsPopular      bool     `json:"is_popular"`
	Features       []string `json:"features"`
	FeatureFlags   map[string]interface{} `json:"feature_flags"`
}

type UpdatePlanRequest struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Price        *float64 `json:"price"`
	IsActive     *bool    `json:"is_active"`
	IsPopular    *bool    `json:"is_popular"`
	SortOrder    *int     `json:"sort_order"`
	JobPostLimit *int     `json:"job_post_limit"`
}

type SubscriptionServiceImpl struct {
	cfg              *config.Config
	subscriptionRepo repository.SubscriptionRepository
	paymentRepo      repository.PaymentRepository
	userRepo         repository.UserRepository
	jobRepo          repository.JobRepository
}

func NewSubscriptionService(
	cfg *config.Config,
	subscriptionRepo repository.SubscriptionRepository,
	paymentRepo repository.PaymentRepository,
	userRepo repository.UserRepository,
	jobRepo repository.JobRepository,
) SubscriptionService {
	return &SubscriptionServiceImpl{
		cfg:              cfg,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		userRepo:         userRepo,
		jobRepo:          jobRepo,
	}
}

// ========== PLAN MANAGEMENT ==========

func (s *SubscriptionServiceImpl) GetPlans(ctx context.Context) ([]*models.SubscriptionPlan, error) {
	return s.subscriptionRepo.GetAllPlans(ctx, false)
}

func (s *SubscriptionServiceImpl) GetPlanByID(ctx context.Context, planID string) (*models.SubscriptionPlan, error) {
	return s.subscriptionRepo.GetPlanByPlanID(ctx, planID)
}

func (s *SubscriptionServiceImpl) CreatePlan(ctx context.Context, req *CreatePlanRequest) (*models.SubscriptionPlan, error) {
	features := make(models.JSONArray, len(req.Features))
	for i, f := range req.Features {
		features[i] = f
	}
	featureFlags := models.JSONMap(req.FeatureFlags)
	if featureFlags == nil {
		featureFlags = make(models.JSONMap)
	}

	plan := &models.SubscriptionPlan{
		PlanID:        req.PlanID,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Currency:      req.Currency,
		Interval:      req.Interval,
		IntervalCount: req.IntervalCount,
		TrialDays:     req.TrialDays,
		JobPostLimit:  req.JobPostLimit,
		SortOrder:     req.SortOrder,
		IsActive:      true,
		IsPopular:     req.IsPopular,
		Features:      features,
		FeatureFlags:  featureFlags,
	}
	
	if plan.Currency == "" {
		plan.Currency = "KES"
	}
	if plan.Interval == "" {
		plan.Interval = "month"
	}
	if plan.IntervalCount == 0 {
		plan.IntervalCount = 1
	}
	
	if err := s.subscriptionRepo.CreatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}
	
	return plan, nil
}

func (s *SubscriptionServiceImpl) UpdatePlan(ctx context.Context, planID string, req *UpdatePlanRequest) (*models.SubscriptionPlan, error) {
	plan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, planID)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}
	if req.IsPopular != nil {
		plan.IsPopular = *req.IsPopular
	}
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}
	if req.JobPostLimit != nil {
		plan.JobPostLimit = *req.JobPostLimit
	}
	
	if err := s.subscriptionRepo.UpdatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}
	
	return plan, nil
}

func (s *SubscriptionServiceImpl) DeletePlan(ctx context.Context, planID string) error {
	plan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, planID)
	if err != nil {
		return err
	}
	
	return s.subscriptionRepo.DeletePlan(ctx, plan.ID)
}

// ========== SUBSCRIPTION MANAGEMENT ==========

func (s *SubscriptionServiceImpl) Subscribe(ctx context.Context, userID string, req *SubscribeRequest) (*models.Subscription, error) {
	plan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	
	_, err = s.paymentRepo.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	
	// Cancel any existing active subscription
	existing, _ := s.subscriptionRepo.GetActiveSubscriptionByUser(ctx, userID)
	if existing != nil {
		s.CancelSubscription(ctx, userID, existing.ID, "New subscription")
	}
	
	now := time.Now()
	periodStart := now
	periodEnd := s.calculatePeriodEnd(now, plan.Interval, plan.IntervalCount)
	
	// Handle trial
	if req.TrialDays > 0 || plan.TrialDays > 0 {
		trialDays := req.TrialDays
		if trialDays == 0 {
			trialDays = plan.TrialDays
		}
		trialEnd := now.AddDate(0, 0, trialDays)
		periodStart = trialEnd
		periodEnd = s.calculatePeriodEnd(trialEnd, plan.Interval, plan.IntervalCount)
	}
	
	subscription := &models.Subscription{
		UserID:             userID,
		PlanID:             plan.PlanID,
		PlanName:           plan.Name,
		Amount:             plan.Price,
		Currency:           plan.Currency,
		Interval:           plan.Interval,
		Status:             "active",
		StartDate:          now,
		EndDate:            periodEnd,
		AutoRenew:          req.AutoRenew,
		JobPostLimit:       plan.JobPostLimit,
		FeatureFlags:       plan.FeatureFlags,
		Features:           plan.Features,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		LastPaymentID:      &req.PaymentID,
	}
	
	if req.TrialDays > 0 || plan.TrialDays > 0 {
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		subscription.TrialEndsAt = &trialEnd
	}
	
	if err := s.subscriptionRepo.CreateSubscription(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	
	// Update employer profile
	updates := map[string]interface{}{
		"subscription_plan":       plan.PlanID,
		"subscription_expires_at": periodEnd,
		"subscription_job_posts":  plan.JobPostLimit,
	}
	s.userRepo.UpdateEmployerProfile(ctx, userID, updates)
	
	// Create subscription usage records
	s.createUsageRecords(ctx, subscription.ID, userID, plan)
	
	// Log history
	history := &models.SubscriptionHistory{
		SubscriptionID: subscription.ID,
		UserID:         userID,
		NewPlanID:      plan.PlanID,
		NewAmount:      plan.Price,
		Action:         "created",
	}
	s.subscriptionRepo.AddHistory(ctx, history)
	
	return subscription, nil
}

func (s *SubscriptionServiceImpl) CancelSubscription(ctx context.Context, userID, subscriptionID string, reason string) error {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}
	
	if sub.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	
	if err := s.subscriptionRepo.CancelSubscription(ctx, subscriptionID, time.Now()); err != nil {
		return err
	}
	
	// Log history
	history := &models.SubscriptionHistory{
		SubscriptionID: subscriptionID,
		UserID:         userID,
		OldPlanID:      sub.PlanID,
		Action:         "cancelled",
		Reason:         reason,
	}
	s.subscriptionRepo.AddHistory(ctx, history)
	
	return nil
}

func (s *SubscriptionServiceImpl) GetCurrentSubscription(ctx context.Context, userID string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetActiveSubscriptionByUser(ctx, userID)
}

func (s *SubscriptionServiceImpl) GetSubscription(ctx context.Context, subscriptionID string) (*models.Subscription, error) {
	return s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
}

func (s *SubscriptionServiceImpl) ListUserSubscriptions(ctx context.Context, userID string, page, limit int) ([]*models.Subscription, int64, error) {
	filters := map[string]interface{}{
		"user_id": userID,
	}
	return s.subscriptionRepo.ListSubscriptions(ctx, filters, page, limit)
}

// ========== FEATURE CHECKING ==========

func (s *SubscriptionServiceImpl) CanAccessFeature(ctx context.Context, userID, feature string) (bool, error) {
	sub, err := s.subscriptionRepo.GetActiveSubscriptionByUser(ctx, userID)
	if err != nil || sub == nil {
		return false, nil
	}
	
	// Check feature flags
	if sub.FeatureFlags != nil {
		if val, ok := sub.FeatureFlags[feature]; ok {
			if enabled, ok := val.(bool); ok {
				return enabled, nil
			}
		}
	}
	
	// Default feature access
	featureMap := map[string][]string{
		"advanced_analytics": {"pro_monthly", "pro_yearly", "business_monthly"},
		"priority_support":   {"pro_monthly", "pro_yearly", "business_monthly"},
		"api_access":         {"business_monthly"},
		"talent_pool":        {"pro_monthly", "pro_yearly", "business_monthly"},
		"featured_jobs":      {"pro_monthly", "pro_yearly", "business_monthly"},
	}
	
	allowedPlans, ok := featureMap[feature]
	if !ok {
		return true, nil
	}
	
	for _, plan := range allowedPlans {
		if sub.PlanID == plan {
			return true, nil
		}
	}
	
	return false, nil
}

func (s *SubscriptionServiceImpl) GetRemainingUsage(ctx context.Context, userID, metricKey string) (int, error) {
	sub, err := s.subscriptionRepo.GetActiveSubscriptionByUser(ctx, userID)
	if err != nil || sub == nil {
		return 0, nil
	}
	
	usage, err := s.subscriptionRepo.GetUsage(ctx, sub.ID, metricKey)
	if err != nil {
		// Create usage record if not exists
		periodStart := sub.CurrentPeriodStart
		periodEnd := sub.CurrentPeriodEnd
		
		limit := s.getLimitForMetric(sub, metricKey)
		usage, err = s.subscriptionRepo.GetOrCreateUsage(ctx, sub.ID, userID, metricKey, limit, periodStart, periodEnd)
		if err != nil {
			return 0, err
		}
	}
	
	remaining := usage.Limit - usage.CurrentUsage
	if remaining < 0 {
		remaining = 0
	}
	
	return remaining, nil
}

func (s *SubscriptionServiceImpl) TrackUsage(ctx context.Context, userID, metricKey string) error {
	sub, err := s.subscriptionRepo.GetActiveSubscriptionByUser(ctx, userID)
	if err != nil || sub == nil {
		return nil
	}
	
	usage, err := s.subscriptionRepo.GetUsage(ctx, sub.ID, metricKey)
	if err != nil {
		limit := s.getLimitForMetric(sub, metricKey)
		usage, err = s.subscriptionRepo.GetOrCreateUsage(ctx, sub.ID, userID, metricKey, limit, sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
		if err != nil {
			return err
		}
	}
	
	// Check if limit exceeded
	if usage.CurrentUsage >= usage.Limit {
		return fmt.Errorf("usage limit exceeded for %s", metricKey)
	}
	
	return s.subscriptionRepo.IncrementUsage(ctx, usage.ID)
}

// ========== BILLING ==========

func (s *SubscriptionServiceImpl) GenerateInvoice(ctx context.Context, subscriptionID, paymentID string) (*models.Invoice, error) {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	
	payment, err := s.paymentRepo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	
	invoiceNumber := fmt.Sprintf("INV-%d-%s", time.Now().Unix(), subscriptionID[:8])
	tax := payment.Amount * 0.16
	total := payment.Amount + tax
	
	invoice := &models.Invoice{
		InvoiceNumber:  invoiceNumber,
		UserID:         sub.UserID,
		SubscriptionID: &subscriptionID,
		PaymentID:      &paymentID,
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		Tax:            tax,
		Total:          total,
		Status:         "paid",
		DueDate:        time.Now().AddDate(0, 0, 14),
		PaidAt:         payment.PaidAt,
	}
	
	if err := s.paymentRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, err
	}
	
	item := &models.InvoiceItem{
		InvoiceID:   invoice.ID,
		Description: fmt.Sprintf("%s Subscription - %s", sub.PlanName, sub.PlanID),
		Quantity:    1,
		UnitPrice:   payment.Amount,
		Total:       payment.Amount,
	}
	s.paymentRepo.CreateInvoiceItem(ctx, item)
	
	return invoice, nil
}

func (s *SubscriptionServiceImpl) GetInvoices(ctx context.Context, userID string, page, limit int) ([]*models.Invoice, int64, error) {
	return s.paymentRepo.ListUserInvoices(ctx, userID, page, limit)
}

// ========== SUBSCRIPTION LIFECYCLE ==========

func (s *SubscriptionServiceImpl) RenewSubscription(ctx context.Context, subscriptionID string) error {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}
	
	if sub.Status != "active" {
		return fmt.Errorf("subscription is not active")
	}
	
	if !sub.AutoRenew {
		return fmt.Errorf("auto-renew is disabled")
	}
	
	// Calculate new period
	newPeriodStart := sub.CurrentPeriodEnd
	newPeriodEnd := s.calculatePeriodEnd(newPeriodStart, sub.Interval, 1)
	
	sub.CurrentPeriodStart = newPeriodStart
	sub.CurrentPeriodEnd = newPeriodEnd
	sub.StartDate = time.Now()
	sub.EndDate = newPeriodEnd
	sub.UpdatedAt = time.Now()
	
	if err := s.subscriptionRepo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}
	
	// Reset usage for new period
	s.subscriptionRepo.ResetUsage(ctx, sub.ID, "job_posts")
	s.subscriptionRepo.ResetUsage(ctx, sub.ID, "api_calls")
	
	// Log history
	history := &models.SubscriptionHistory{
		SubscriptionID: sub.ID,
		UserID:         sub.UserID,
		NewPlanID:      sub.PlanID,
		NewAmount:      sub.Amount,
		Action:         "renewed",
	}
	s.subscriptionRepo.AddHistory(ctx, history)
	
	return nil
}

func (s *SubscriptionServiceImpl) ExpireSubscriptions(ctx context.Context) error {
	expired, err := s.subscriptionRepo.GetExpiredSubscriptions(ctx)
	if err != nil {
		return err
	}
	
	for _, sub := range expired {
		sub.Status = "expired"
		s.subscriptionRepo.UpdateSubscription(ctx, sub)
		
		history := &models.SubscriptionHistory{
			SubscriptionID: sub.ID,
			UserID:         sub.UserID,
			Action:         "expired",
		}
		s.subscriptionRepo.AddHistory(ctx, history)
	}
	
	return nil
}

func (s *SubscriptionServiceImpl) UpgradeSubscription(ctx context.Context, userID, subscriptionID, newPlanID string) (*models.Subscription, error) {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	
	if sub.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	
	newPlan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, newPlanID)
	if err != nil {
		return nil, err
	}
	
	oldPlanID := sub.PlanID
	oldAmount := sub.Amount
	
	// Update subscription
	sub.PlanID = newPlan.PlanID
	sub.PlanName = newPlan.Name
	sub.Amount = newPlan.Price
	sub.JobPostLimit = newPlan.JobPostLimit
	sub.FeatureFlags = newPlan.FeatureFlags
	sub.Features = newPlan.Features
	sub.UpdatedAt = time.Now()
	
	if err := s.subscriptionRepo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	
	// Update employer profile
	updates := map[string]interface{}{
		"subscription_plan":      newPlan.PlanID,
		"subscription_job_posts": newPlan.JobPostLimit,
	}
	s.userRepo.UpdateEmployerProfile(ctx, userID, updates)
	
	// Log history
	history := &models.SubscriptionHistory{
		SubscriptionID: sub.ID,
		UserID:         userID,
		OldPlanID:      oldPlanID,
		NewPlanID:      newPlan.PlanID,
		OldAmount:      oldAmount,
		NewAmount:      newPlan.Price,
		Action:         "upgraded",
	}
	s.subscriptionRepo.AddHistory(ctx, history)
	
	return sub, nil
}

func (s *SubscriptionServiceImpl) DowngradeSubscription(ctx context.Context, userID, subscriptionID, newPlanID string) (*models.Subscription, error) {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	
	if sub.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	
	newPlan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, newPlanID)
	if err != nil {
		return nil, err
	}
	
	oldPlanID := sub.PlanID
	oldAmount := sub.Amount
	
	sub.PlanID = newPlan.PlanID
	sub.PlanName = newPlan.Name
	sub.Amount = newPlan.Price
	sub.JobPostLimit = newPlan.JobPostLimit
	sub.FeatureFlags = newPlan.FeatureFlags
	sub.Features = newPlan.Features
	sub.UpdatedAt = time.Now()
	
	if err := s.subscriptionRepo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}
	
	updates := map[string]interface{}{
		"subscription_plan":      newPlan.PlanID,
		"subscription_job_posts": newPlan.JobPostLimit,
	}
	s.userRepo.UpdateEmployerProfile(ctx, userID, updates)
	
	history := &models.SubscriptionHistory{
		SubscriptionID: sub.ID,
		UserID:         userID,
		OldPlanID:      oldPlanID,
		NewPlanID:      newPlan.PlanID,
		OldAmount:      oldAmount,
		NewAmount:      newPlan.Price,
		Action:         "downgraded",
	}
	s.subscriptionRepo.AddHistory(ctx, history)
	
	return sub, nil
}

// ========== ADMIN ==========

func (s *SubscriptionServiceImpl) GetAllSubscriptions(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Subscription, int64, error) {
	return s.subscriptionRepo.ListSubscriptions(ctx, filters, page, limit)
}

func (s *SubscriptionServiceImpl) GetSubscriptionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]*models.SubscriptionHistory, int64, error) {
	return s.subscriptionRepo.GetSubscriptionHistory(ctx, subscriptionID, page, limit)
}

// ========== HELPER METHODS ==========

func (s *SubscriptionServiceImpl) calculatePeriodEnd(start time.Time, interval string, intervalCount int) time.Time {
	if interval == "month" {
		return start.AddDate(0, intervalCount, 0)
	}
	return start.AddDate(intervalCount, 0, 0)
}

func (s *SubscriptionServiceImpl) getLimitForMetric(sub *models.Subscription, metricKey string) int {
	switch metricKey {
	case "job_posts":
		return sub.JobPostLimit
	case "api_calls":
		if sub.PlanID == "business_monthly" {
			return 10000
		}
		return 1000
	default:
		return 0
	}
}

func (s *SubscriptionServiceImpl) createUsageRecords(ctx context.Context, subscriptionID, userID string, plan *models.SubscriptionPlan) {
	metrics := []struct {
		key   string
		limit int
	}{
		{"job_posts", plan.JobPostLimit},
		{"api_calls", 1000},
	}
	
	for _, m := range metrics {
		s.subscriptionRepo.GetOrCreateUsage(ctx, subscriptionID, userID, m.key, m.limit, time.Now(), time.Now())
	}
}