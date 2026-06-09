package services

import (
	"context"
	"fmt"
	"time"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type PaymentRetryState string

const (
	PaymentRetryNone    PaymentRetryState = ""
	PaymentRetryFirst   PaymentRetryState = "first_retry"
	PaymentRetrySecond  PaymentRetryState = "second_retry"
	PaymentRetryFinal   PaymentRetryState = "final_retry"
)

type AddPaymentMethodRequest struct {
	Type        string `json:"type" validate:"required"`        // mpesa, card, bank
	PhoneNumber string `json:"phone_number"`
	CardToken   string `json:"card_token"`
	SetDefault  bool   `json:"set_default"`
}

type CalculateProrationResult struct {
	CreditAmount float64 `json:"credit_amount"`
	NewAmount    float64 `json:"new_amount"`
	DueNow       float64 `json:"due_now"`
	DaysLeft     int     `json:"days_left"`
	TotalDays    int     `json:"total_days"`
}

type SubscriptionService interface {
	// Plan management
	GetPlans(ctx context.Context) ([]*models.SubscriptionPlan, error)
	GetAllPlans(ctx context.Context, includeInactive bool) ([]*models.SubscriptionPlan, error)
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
	ReactivateSubscription(ctx context.Context, userID, subscriptionID string) (*models.Subscription, error)
	
	// Payment methods
	GetPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethod, error)
	AddPaymentMethod(ctx context.Context, userID string, req *AddPaymentMethodRequest) (*models.PaymentMethod, error)
	RemovePaymentMethod(ctx context.Context, userID, methodID string) error
	SetDefaultPaymentMethod(ctx context.Context, userID, methodID string) error
	
	// Payment processing
	SubscribeWithNewPayment(ctx context.Context, userID, planID, phoneNumber string) (*models.Subscription, *IntaSendChargeResponse, error)
	VerifyPayment(ctx context.Context, transactionID, reference string) (*models.Payment, error)
	RetryPayment(ctx context.Context, subscriptionID string) error
	ProcessPaymentRetries(ctx context.Context) error
	
	// Webhook
	HandleWebhook(ctx context.Context, payload *models.IntaSendWebhookPayload) error

	// Proration
	CalculateProration(ctx context.Context, subscriptionID, newPlanID string) (*CalculateProrationResult, error)
	
	// Invoice
	GenerateInvoicePDF(ctx context.Context, invoiceID string) (string, error)
	
	// Plans
	GetPlanFeatures(ctx context.Context, planID string) ([]*models.SubscriptionPlanFeature, error)
	
	// Admin
	GetAllSubscriptions(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Subscription, int64, error)
	GetSubscriptionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]*models.SubscriptionHistory, int64, error)
	AdminAssignSubscription(ctx context.Context, userID, planID string) (*models.Subscription, error)
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
	intaSend         *IntaSendClient
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
		intaSend:         NewIntaSendClient(cfg),
	}
}

// ========== PLAN MANAGEMENT ==========

func (s *SubscriptionServiceImpl) GetPlans(ctx context.Context) ([]*models.SubscriptionPlan, error) {
	return s.subscriptionRepo.GetAllPlans(ctx, false)
}

func (s *SubscriptionServiceImpl) GetAllPlans(ctx context.Context, includeInactive bool) ([]*models.SubscriptionPlan, error) {
	return s.subscriptionRepo.GetAllPlans(ctx, includeInactive)
}

// GetPlanByID retrieves a plan by ID (supports both UUID and plan_id)
func (s *SubscriptionServiceImpl) GetPlanByID(ctx context.Context, id string) (*models.SubscriptionPlan, error) {
	// Try by plan_id first (string key like "free", "pro_monthly")
	plan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, id)
	if err == nil {
		return plan, nil
	}
	
	// If not found, try by UUID
	return s.subscriptionRepo.GetPlan(ctx, id)
}


func (s *SubscriptionServiceImpl) CreatePlan(ctx context.Context, req *CreatePlanRequest) (*models.SubscriptionPlan, error) {
	features := make(models.StringArray, len(req.Features))
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
		// Trial expired — downgrade to free plan instead of marking expired
		if sub.Status == "trialing" {
			freePlan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, "free")
			if err != nil {
				// If free plan doesn't exist, just expire it
				sub.Status = "expired"
				s.subscriptionRepo.UpdateSubscription(ctx, sub)
				continue
			}
			
			sub.Status = "active"
			sub.PlanID = freePlan.PlanID
			sub.PlanName = freePlan.Name
			sub.Amount = freePlan.Price
			sub.JobPostLimit = freePlan.JobPostLimit
			sub.Features = freePlan.Features
			sub.FeatureFlags = freePlan.FeatureFlags
			// Set end_date far in the future since free plans don't expire
			sub.EndDate = time.Now().AddDate(100, 0, 0)
			sub.CurrentPeriodEnd = time.Now().AddDate(100, 0, 0)
			sub.TrialEndsAt = nil
			sub.AutoRenew = false
			sub.UpdatedAt = time.Now()
			s.subscriptionRepo.UpdateSubscription(ctx, sub)
			
			history := &models.SubscriptionHistory{
				SubscriptionID: sub.ID,
				UserID:         sub.UserID,
				OldPlanID:      sub.PlanID,
				NewPlanID:      freePlan.PlanID,
				Action:         "downgraded_from_trial",
				Reason:         "Trial period ended",
			}
			s.subscriptionRepo.AddHistory(ctx, history)

			// Update employer profile's subscription_plan field
			s.userRepo.UpdateEmployerProfile(ctx, sub.UserID, map[string]interface{}{
				"subscription_plan": "free",
			})
			continue
		}
		
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

// ========== REACTIVATE ==========

func (s *SubscriptionServiceImpl) ReactivateSubscription(ctx context.Context, userID, subscriptionID string) (*models.Subscription, error) {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	if sub.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	if sub.Status != "cancelled" && sub.Status != "expired" {
		return nil, fmt.Errorf("subscription is not cancelled or expired")
	}

	now := time.Now()
	newPeriodEnd := s.calculatePeriodEnd(now, sub.Interval, 1)

	sub.Status = "active"
	sub.CancelledAt = nil
	sub.StartDate = now
	sub.EndDate = newPeriodEnd
	sub.CurrentPeriodStart = now
	sub.CurrentPeriodEnd = newPeriodEnd
	sub.AutoRenew = true
	sub.UpdatedAt = now

	if err := s.subscriptionRepo.UpdateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	history := &models.SubscriptionHistory{
		SubscriptionID: sub.ID,
		UserID:         userID,
		NewPlanID:      sub.PlanID,
		NewAmount:      sub.Amount,
		Action:         "reactivated",
	}
	s.subscriptionRepo.AddHistory(ctx, history)

	return sub, nil
}

// ========== PAYMENT METHODS ==========

func (s *SubscriptionServiceImpl) GetPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethod, error) {
	return s.paymentRepo.ListUserPaymentMethods(ctx, userID)
}

func (s *SubscriptionServiceImpl) AddPaymentMethod(ctx context.Context, userID string, req *AddPaymentMethodRequest) (*models.PaymentMethod, error) {
	pm := &models.PaymentMethod{
		UserID:      userID,
		Type:        req.Type,
		PhoneNumber: req.PhoneNumber,
		IsDefault:   req.SetDefault,
		IsValid:     true,
	}

	if err := s.paymentRepo.CreatePaymentMethod(ctx, pm); err != nil {
		return nil, err
	}

	if req.SetDefault {
		s.paymentRepo.SetDefaultPaymentMethod(ctx, userID, pm.ID)
	}

	return pm, nil
}

func (s *SubscriptionServiceImpl) RemovePaymentMethod(ctx context.Context, userID, methodID string) error {
	pm, err := s.paymentRepo.GetPaymentMethod(ctx, methodID)
	if err != nil {
		return err
	}
	if pm.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.paymentRepo.DeletePaymentMethod(ctx, methodID)
}

func (s *SubscriptionServiceImpl) SetDefaultPaymentMethod(ctx context.Context, userID, methodID string) error {
	return s.paymentRepo.SetDefaultPaymentMethod(ctx, userID, methodID)
}

// ========== PAYMENT PROCESSING ==========

func (s *SubscriptionServiceImpl) SubscribeWithNewPayment(ctx context.Context, userID, planID, phoneNumber string) (*models.Subscription, *IntaSendChargeResponse, error) {
	plan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, planID)
	if err != nil {
		return nil, nil, fmt.Errorf("plan not found: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}

	reference := fmt.Sprintf("SUB-%s-%d", userID[:8], time.Now().Unix())
	chargeReq := &IntaSendChargeRequest{
		Amount:      plan.Price,
		Currency:    plan.Currency,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		Reference:   reference,
		Narrative:   fmt.Sprintf("%s %s Subscription", plan.Name, plan.Interval),
		WebhookURL:  s.cfg.AppURL + "/api/v1/payments/webhook",
		RedirectURL: s.cfg.AppURL + "/subscriptions/success",
	}

	chargeResp, err := s.intaSend.CreateCharge(chargeReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create charge: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)
	intent := &models.PaymentIntent{
		UserID:      userID,
		Amount:      plan.Price,
		Currency:    plan.Currency,
		PlanID:      plan.PlanID,
		Status:      "pending",
		InvoiceID:   chargeResp.InvoiceID,
		RedirectURL: chargeResp.RedirectURL,
		ExpiresAt:   expiresAt,
	}
	s.paymentRepo.CreatePaymentIntent(ctx, intent)

	payment := &models.Payment{
		UserID:        userID,
		Amount:        plan.Price,
		Currency:      plan.Currency,
		Status:        "pending",
		PaymentMethod: "mpesa",
		Reference:     reference,
		Description:   fmt.Sprintf("%s %s - %s", plan.Name, plan.Interval, plan.PlanID),
	}
	s.paymentRepo.CreatePayment(ctx, payment)

	var subscription *models.Subscription
	if chargeResp.Status == "completed" || chargeResp.Status == "success" {
		subscription, err = s.Subscribe(ctx, userID, &SubscribeRequest{
			PlanID:    plan.PlanID,
			PaymentID: payment.ID,
			AutoRenew: true,
		})
		if err != nil {
			return nil, chargeResp, fmt.Errorf("subscription creation failed: %w", err)
		}
	}

	return subscription, chargeResp, nil
}

func (s *SubscriptionServiceImpl) VerifyPayment(ctx context.Context, transactionID, reference string) (*models.Payment, error) {
	if transactionID != "" {
		payment, err := s.paymentRepo.GetPaymentByTransactionID(ctx, transactionID)
		if err == nil {
			return payment, nil
		}

		statusResp, err := s.intaSend.GetPaymentStatus(transactionID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify payment: %w", err)
		}

		now := time.Now()
		if statusResp.Status == "completed" || statusResp.Status == "success" {
			if err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.ID, "completed", transactionID, &now); err != nil {
				return nil, err
			}
			payment.Status = "completed"
			payment.TransactionID = transactionID
			payment.PaidAt = &now
		}
		return payment, nil
	}

	if reference != "" {
		payment, err := s.paymentRepo.GetPaymentByReference(ctx, reference)
		if err != nil {
			return nil, err
		}
		return payment, nil
	}

	return nil, fmt.Errorf("either transaction_id or reference is required")
}

func (s *SubscriptionServiceImpl) RetryPayment(ctx context.Context, subscriptionID string) error {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	if sub.Status != "past_due" {
		return fmt.Errorf("subscription is not past_due")
	}

	user, err := s.userRepo.GetByID(ctx, sub.UserID)
	if err != nil {
		return err
	}

	pm, err := s.paymentRepo.GetDefaultPaymentMethod(ctx, sub.UserID)
	if err != nil {
		return fmt.Errorf("no default payment method: %w", err)
	}

	reference := fmt.Sprintf("RETRY-%s-%d", sub.ID[:8], time.Now().Unix())
	phoneNumber := pm.PhoneNumber

	chargeReq := &IntaSendChargeRequest{
		Amount:      sub.Amount,
		Currency:    sub.Currency,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		Reference:   reference,
		Narrative:   fmt.Sprintf("Retry payment - %s", sub.PlanName),
		WebhookURL:  s.cfg.AppURL + "/api/v1/payments/webhook",
		RedirectURL: s.cfg.AppURL + "/subscriptions/success",
	}

	_, err = s.intaSend.CreateCharge(chargeReq)
	if err != nil {
		return fmt.Errorf("retry charge failed: %w", err)
	}

	payment := &models.Payment{
		UserID:        sub.UserID,
		SubscriptionID: &sub.ID,
		Amount:        sub.Amount,
		Currency:      sub.Currency,
		Status:        "pending",
		PaymentMethod: string(pm.Type),
		Reference:     reference,
		Description:   fmt.Sprintf("Retry: %s Subscription", sub.PlanName),
	}
	return s.paymentRepo.CreatePayment(ctx, payment)
}

func (s *SubscriptionServiceImpl) ProcessPaymentRetries(ctx context.Context) error {
	expiring, err := s.subscriptionRepo.GetExpiringSoonSubscriptions(ctx, 1)
	if err != nil {
		return err
	}

	for _, sub := range expiring {
		s.RetryPayment(ctx, sub.ID)
	}

	history, _, _ := s.subscriptionRepo.GetSubscriptionHistory(ctx, "", 1, 1000)
	for _, h := range history {
		if h.Action == "payment_failed" {
			sub, err := s.subscriptionRepo.GetSubscription(ctx, h.SubscriptionID)
			if err != nil || sub.Status != "past_due" {
				continue
			}
			s.RetryPayment(ctx, sub.ID)
		}
	}

	return nil
}

// ========== WEBHOOK ==========

func (s *SubscriptionServiceImpl) HandleWebhook(ctx context.Context, payload *models.IntaSendWebhookPayload) error {
	switch payload.Event {
	case "payment.completed":
		return s.handlePaymentCompleted(ctx, payload.Data)
	case "payment.failed":
		return s.handlePaymentFailed(ctx, payload.Data)
	case "refund.completed":
		return s.handleRefundCompleted(ctx, payload.Data)
	default:
		return fmt.Errorf("unknown webhook event: %s", payload.Event)
	}
}

func (s *SubscriptionServiceImpl) handlePaymentCompleted(ctx context.Context, data map[string]interface{}) error {
	transactionID, _ := data["transaction_id"].(string)
	reference, _ := data["reference"].(string)

	if reference == "" || transactionID == "" {
		return fmt.Errorf("missing reference or transaction_id")
	}

	payment, err := s.paymentRepo.GetPaymentByReference(ctx, reference)
	if err != nil {
		return fmt.Errorf("payment not found for reference %s: %w", reference, err)
	}

	now := time.Now()
	if err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.ID, "completed", transactionID, &now); err != nil {
		return err
	}

	invoice, err := s.GenerateInvoice(ctx, payment.ID, payment.ID)
	if err != nil {
		return fmt.Errorf("invoice generation failed: %w", err)
	}

	if payment.SubscriptionID != nil {
		sub, _ := s.subscriptionRepo.GetSubscription(ctx, *payment.SubscriptionID)
		if sub != nil && sub.Status == "past_due" {
			sub.Status = "active"
			s.subscriptionRepo.UpdateSubscription(ctx, sub)
		}
	}

	_ = invoice
	return nil
}

func (s *SubscriptionServiceImpl) handlePaymentFailed(ctx context.Context, data map[string]interface{}) error {
	reference, _ := data["reference"].(string)

	payment, err := s.paymentRepo.GetPaymentByReference(ctx, reference)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	now := time.Now()
	s.paymentRepo.UpdatePaymentStatus(ctx, payment.ID, "failed", "", &now)

	if payment.SubscriptionID != nil {
		sub, err := s.subscriptionRepo.GetSubscription(ctx, *payment.SubscriptionID)
		if err == nil && sub.Status == "active" {
			sub.Status = "past_due"
			sub.UpdatedAt = time.Now()
			s.subscriptionRepo.UpdateSubscription(ctx, sub)

			history := &models.SubscriptionHistory{
				SubscriptionID: sub.ID,
				UserID:         sub.UserID,
				Action:         "payment_failed",
				Reason:         "Payment gateway declined transaction",
			}
			s.subscriptionRepo.AddHistory(ctx, history)
		}
	}

	return nil
}

func (s *SubscriptionServiceImpl) handleRefundCompleted(ctx context.Context, data map[string]interface{}) error {
	transactionID, _ := data["transaction_id"].(string)

	payment, err := s.paymentRepo.GetPaymentByTransactionID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	now := time.Now()
	return s.paymentRepo.UpdatePaymentStatus(ctx, payment.ID, "refunded", transactionID, &now)
}

// ========== PRORATION ==========

func (s *SubscriptionServiceImpl) CalculateProration(ctx context.Context, subscriptionID, newPlanID string) (*CalculateProrationResult, error) {
	sub, err := s.subscriptionRepo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	newPlan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, newPlanID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	totalDays := int(sub.CurrentPeriodEnd.Sub(sub.CurrentPeriodStart).Hours() / 24)
	daysLeft := int(sub.CurrentPeriodEnd.Sub(now).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}
	if totalDays <= 0 {
		totalDays = 30
	}

	dailyRate := sub.Amount / float64(totalDays)
	creditAmount := dailyRate * float64(daysLeft)

	newDailyRate := newPlan.Price / float64(totalDays)
	newAmount := newDailyRate * float64(daysLeft)

	dueNow := newAmount - creditAmount
	if dueNow < 0 {
		dueNow = 0
	}

	return &CalculateProrationResult{
		CreditAmount: creditAmount,
		NewAmount:    newAmount,
		DueNow:       dueNow,
		DaysLeft:     daysLeft,
		TotalDays:    totalDays,
	}, nil
}

// ========== INVOICE PDF ==========

func (s *SubscriptionServiceImpl) GenerateInvoicePDF(ctx context.Context, invoiceID string) (string, error) {
	invoice, err := s.paymentRepo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return "", err
	}

	items, err := s.paymentRepo.GetInvoiceItems(ctx, invoiceID)
	if err != nil {
		return "", err
	}

	pdfURL := fmt.Sprintf("%s/invoices/%s/view", s.cfg.AppURL, invoice.ID)

	s.paymentRepo.UpdateInvoicePDF(ctx, invoiceID, pdfURL)

	_ = items
	return pdfURL, nil
}

// ========== PLANS ==========

func (s *SubscriptionServiceImpl) GetPlanFeatures(ctx context.Context, planID string) ([]*models.SubscriptionPlanFeature, error) {
	return s.subscriptionRepo.GetPlanFeatures(ctx, planID)
}

// ========== ADMIN ==========

func (s *SubscriptionServiceImpl) GetAllSubscriptions(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Subscription, int64, error) {
	return s.subscriptionRepo.ListSubscriptions(ctx, filters, page, limit)
}

func (s *SubscriptionServiceImpl) GetSubscriptionHistory(ctx context.Context, subscriptionID string, page, limit int) ([]*models.SubscriptionHistory, int64, error) {
	return s.subscriptionRepo.GetSubscriptionHistory(ctx, subscriptionID, page, limit)
}

func (s *SubscriptionServiceImpl) AdminAssignSubscription(ctx context.Context, userID, planID string) (*models.Subscription, error) {
	plan, err := s.subscriptionRepo.GetPlanByPlanID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Cancel any existing active subscription
	existing, _ := s.subscriptionRepo.GetActiveSubscriptionByUser(ctx, userID)
	if existing != nil {
		s.CancelSubscription(ctx, userID, existing.ID, "Admin reassigned to "+plan.PlanID)
	}

	now := time.Now()
	periodEnd := s.calculatePeriodEnd(now, plan.Interval, plan.IntervalCount)

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
		AutoRenew:          true,
		JobPostLimit:       plan.JobPostLimit,
		FeatureFlags:       plan.FeatureFlags,
		Features:           plan.Features,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	if plan.TrialDays > 0 {
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
		Action:         "admin_assigned",
	}
	s.subscriptionRepo.AddHistory(ctx, history)

	return subscription, nil
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