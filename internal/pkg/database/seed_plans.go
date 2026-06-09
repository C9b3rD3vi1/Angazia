package database

import (
	"context"
	"log"

	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func SeedSubscriptionPlans(ctx context.Context, planService services.SubscriptionService) error {
	log.Println("🌱 Seeding subscription plans...")

	plans := []services.CreatePlanRequest{
		// Free Plan
		{
			PlanID:        "free",
			Name:          "Free",
			Description:   "Perfect for getting started with basic job postings",
			Price:         0,
			Currency:      "KES",
			Interval:      "month",
			IntervalCount: 1,
			TrialDays:     0,
			JobPostLimit:  3,
			SortOrder:     1,
			IsPopular:     false,
			Features: []string{
				"3 active job posts",
				"Basic job matching",
				"Email support",
				"Basic analytics",
				"Standard job listing",
			},
			FeatureFlags: map[string]interface{}{
				"advanced_analytics": false,
				"priority_support":   false,
				"api_access":         false,
				"talent_pool":        false,
				"featured_jobs":      false,
				"custom_branding":    false,
				"dedicated_manager":  false,
			},
		},
		// Pro Monthly Plan
		{
			PlanID:        "pro_monthly",
			Name:          "Professional",
			Description:   "For growing companies that need more features",
			Price:         2500,
			Currency:      "KES",
			Interval:      "month",
			IntervalCount: 1,
			TrialDays:     14,
			JobPostLimit:  20,
			SortOrder:     2,
			IsPopular:     true,
			Features: []string{
				"20 active job posts",
				"Advanced AI matching",
				"Priority email support",
				"Advanced analytics dashboard",
				"Featured job listings",
				"Talent pool access",
				"Applicant tracking system",
				"Custom job alerts",
			},
			FeatureFlags: map[string]interface{}{
				"advanced_analytics": true,
				"priority_support":   true,
				"api_access":         false,
				"talent_pool":        true,
				"featured_jobs":      true,
				"custom_branding":    false,
				"dedicated_manager":  false,
			},
		},
		// Pro Yearly Plan
		{
			PlanID:        "pro_yearly",
			Name:          "Professional (Yearly)",
			Description:   "Save 20% with annual billing",
			Price:         24000,
			Currency:      "KES",
			Interval:      "year",
			IntervalCount: 1,
			TrialDays:     14,
			JobPostLimit:  20,
			SortOrder:     3,
			IsPopular:     false,
			Features: []string{
				"20 active job posts",
				"Advanced AI matching",
				"Priority email support",
				"Advanced analytics dashboard",
				"Featured job listings",
				"Talent pool access",
				"Applicant tracking system",
				"Custom job alerts",
				"2 months free",
			},
			FeatureFlags: map[string]interface{}{
				"advanced_analytics": true,
				"priority_support":   true,
				"api_access":         false,
				"talent_pool":        true,
				"featured_jobs":      true,
				"custom_branding":    false,
				"dedicated_manager":  false,
			},
		},
		// Business Monthly Plan
		{
			PlanID:        "business_monthly",
			Name:          "Business",
			Description:   "For scaling teams and enterprises",
			Price:         7500,
			Currency:      "KES",
			Interval:      "month",
			IntervalCount: 1,
			TrialDays:     14,
			JobPostLimit:  100,
			SortOrder:     4,
			IsPopular:     false,
			Features: []string{
				"100 active job posts",
				"Unlimited AI matching",
				"24/7 priority phone support",
				"Custom analytics and reporting",
				"Full API access",
				"Dedicated account manager",
				"Custom branding on job posts",
				"Bulk applicant export",
				"Custom interview workflows",
				"Team collaboration tools",
			},
			FeatureFlags: map[string]interface{}{
				"advanced_analytics": true,
				"priority_support":   true,
				"api_access":         true,
				"talent_pool":        true,
				"featured_jobs":      true,
				"custom_branding":    true,
				"dedicated_manager":  true,
			},
		},
		// Business Yearly Plan
		{
			PlanID:        "business_yearly",
			Name:          "Business (Yearly)",
			Description:   "Save 20% with annual billing for enterprises",
			Price:         72000,
			Currency:      "KES",
			Interval:      "year",
			IntervalCount: 1,
			TrialDays:     14,
			JobPostLimit:  100,
			SortOrder:     5,
			IsPopular:     false,
			Features: []string{
				"100 active job posts",
				"Unlimited AI matching",
				"24/7 priority phone support",
				"Custom analytics and reporting",
				"Full API access",
				"Dedicated account manager",
				"Custom branding on job posts",
				"Bulk applicant export",
				"Custom interview workflows",
				"Team collaboration tools",
				"2 months free",
			},
			FeatureFlags: map[string]interface{}{
				"advanced_analytics": true,
				"priority_support":   true,
				"api_access":         true,
				"talent_pool":        true,
				"featured_jobs":      true,
				"custom_branding":    true,
				"dedicated_manager":  true,
			},
		},
		// Enterprise Plan
		{
			PlanID:        "enterprise",
			Name:          "Enterprise",
			Description:   "Custom solution for large organizations",
			Price:         0,
			Currency:      "KES",
			Interval:      "year",
			IntervalCount: 1,
			TrialDays:     30,
			JobPostLimit:  1000,
			SortOrder:     6,
			IsPopular:     false,
			Features: []string{
				"Unlimited job posts",
				"Custom AI model training",
				"Dedicated support team",
				"SLA guarantees",
				"On-premise deployment option",
				"SSO integration",
				"Custom reporting",
				"White-label solution",
				"Volume discounts",
				"Training sessions",
			},
			FeatureFlags: map[string]interface{}{
				"advanced_analytics": true,
				"priority_support":   true,
				"api_access":         true,
				"talent_pool":        true,
				"featured_jobs":      true,
				"custom_branding":    true,
				"dedicated_manager":  true,
			},
		},
	}

	for _, planReq := range plans {
		// Check if plan already exists
		existing, _ := planService.GetPlanByID(ctx, planReq.PlanID)
		if existing != nil {
			log.Printf("Plan %s already exists, skipping...", planReq.PlanID)
			continue
		}
		
		log.Printf("Creating plan %s...", planReq.PlanID)
		if _, err := planService.CreatePlan(ctx, &planReq); err != nil {
			log.Printf("Failed to create plan %s: %v", planReq.PlanID, err)
		} else {
			log.Printf("✅ Created plan: %s", planReq.PlanID)
		}
	}

	log.Println("✅ Subscription plans seeded successfully")
	return nil
}