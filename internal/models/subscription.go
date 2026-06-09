package models

import (
	"time"
)


// Subscription represents a user's subscription to a plan
type Subscription struct {
	ID                  string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID              string     `json:"user_id" gorm:"type:uuid;not null;index"`
	PlanID              string     `json:"plan_id" gorm:"size:100;not null;index"`
	PlanName            string     `json:"plan_name" gorm:"size:255;not null"`
	Amount              float64    `json:"amount" gorm:"not null"`
	Currency            string     `json:"currency" gorm:"size:3;default:'KES'"`
	Interval            string     `json:"interval" gorm:"size:20;default:'month'"`
	Status              string     `json:"status" gorm:"size:50;default:'active';index"`
	StartDate           time.Time  `json:"start_date" gorm:"not null;index"`
	EndDate             time.Time  `json:"end_date" gorm:"not null;index"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	AutoRenew           bool       `json:"auto_renew" gorm:"default:true"`
	JobPostLimit        int        `json:"job_post_limit" gorm:"default:3"`
	FeatureFlags        JSONMap    `json:"feature_flags" gorm:"type:jsonb"`
	Features            StringArray  `json:"features" gorm:"type:jsonb"`
	CurrentPeriodStart  time.Time  `json:"current_period_start" gorm:"not null"`
	CurrentPeriodEnd    time.Time  `json:"current_period_end" gorm:"not null"`
	TrialEndsAt         *time.Time `json:"trial_ends_at,omitempty"`
	LastPaymentID       *string    `json:"last_payment_id,omitempty" gorm:"type:uuid;index"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}


// SubscriptionPlan represents a available subscription plan
type SubscriptionPlan struct {
	ID            string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlanID        string     `json:"plan_id" gorm:"size:100;uniqueIndex;not null"`
	Name          string     `json:"name" gorm:"size:255;not null"`
	Description   string     `json:"description" gorm:"type:text"`
	Price         float64    `json:"price" gorm:"not null"`
	Currency      string     `json:"currency" gorm:"size:3;default:'KES'"`
	Interval      string     `json:"interval" gorm:"size:20;default:'month'"`
	IntervalCount int        `json:"interval_count" gorm:"default:1"`
	TrialDays     int        `json:"trial_days" gorm:"default:0"`
	JobPostLimit  int        `json:"job_post_limit" gorm:"default:3"`
	SortOrder     int        `json:"sort_order" gorm:"default:0"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	IsPopular     bool       `json:"is_popular" gorm:"default:false"`
	Metadata      JSONMap    `json:"metadata" gorm:"type:jsonb"`
	Features      StringArray `json:"features" gorm:"type:jsonb"`
	FeatureFlags  JSONMap    `json:"feature_flags" gorm:"type:jsonb"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// SubscriptionHistory tracks subscription changes
type SubscriptionHistory struct {
	ID             string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SubscriptionID string     `json:"subscription_id" gorm:"type:uuid;not null;index"`
	UserID         string     `json:"user_id" gorm:"type:uuid;not null;index"`
	OldPlanID      string     `json:"old_plan_id" gorm:"size:100"`
	NewPlanID      string     `json:"new_plan_id" gorm:"size:100;not null"`
	OldAmount      float64    `json:"old_amount"`
	NewAmount      float64    `json:"new_amount"`
	Action         string     `json:"action" gorm:"size:50;not null"`
	Reason         string     `json:"reason" gorm:"type:text"`
	Metadata       JSONMap    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// SubscriptionPlanFeature represents a feature mapping to plan
type SubscriptionPlanFeature struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlanID       string    `json:"plan_id" gorm:"type:uuid;not null;index"`
	FeatureKey   string    `json:"feature_key" gorm:"size:100;not null"`
	FeatureValue string    `json:"feature_value" gorm:"type:text"`
	IsEnabled    bool      `json:"is_enabled" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SubscriptionUsage tracks usage against limits
type SubscriptionUsage struct {
	ID             string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SubscriptionID string     `json:"subscription_id" gorm:"type:uuid;not null;index"`
	UserID         string     `json:"user_id" gorm:"type:uuid;not null;index"`
	MetricKey      string     `json:"metric_key" gorm:"size:100;not null;index"`
	CurrentUsage   int        `json:"current_usage" gorm:"default:0"`
	Limit          int        `json:"limit" gorm:"default:0"`
	PeriodStart    time.Time  `json:"period_start"`
	PeriodEnd      time.Time  `json:"period_end"`
	ResetAt        *time.Time `json:"reset_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

func (SubscriptionHistory) TableName() string {
	return "subscription_history"
}

func (SubscriptionPlanFeature) TableName() string {
	return "subscription_plan_features"
}

func (SubscriptionUsage) TableName() string {
	return "subscription_usage"
}