package database

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

// BackfillEmployerSubscriptions creates subscription records for existing employers
// who registered before the auto-subscription-on-registration feature was added.
func BackfillEmployerSubscriptions(ctx context.Context, db *gorm.DB) error {
	log.Println("🔄 Backfilling subscriptions for existing employers...")

	var freePlan models.SubscriptionPlan
	if err := db.WithContext(ctx).Where("plan_id = ?", "free").First(&freePlan).Error; err != nil {
		return err
	}

	var count int64
	db.WithContext(ctx).Model(&models.User{}).
		Joins("LEFT JOIN subscriptions ON subscriptions.user_id = users.id").
		Where("users.role = ? AND subscriptions.id IS NULL", "employer").
		Count(&count)
	log.Printf("📝 Found %d employers without subscriptions", count)

	var users []models.User
	err := db.WithContext(ctx).
		Joins("LEFT JOIN subscriptions ON subscriptions.user_id = users.id").
		Where("users.role = ? AND subscriptions.id IS NULL", "employer").
		Find(&users).Error
	if err != nil {
		return err
	}

	now := time.Now()
	farFuture := now.AddDate(100, 0, 0)

	for _, user := range users {
		sub := &models.Subscription{
			ID:                uuid.New().String(),
			UserID:            user.ID,
			PlanID:            freePlan.PlanID,
			PlanName:          freePlan.Name,
			Amount:            freePlan.Price,
			Currency:          freePlan.Currency,
			Interval:          freePlan.Interval,
			Status:            "active",
			StartDate:         now,
			EndDate:           farFuture,
			AutoRenew:         false,
			JobPostLimit:      freePlan.JobPostLimit,
			Features:          freePlan.Features,
			FeatureFlags:      freePlan.FeatureFlags,
			CurrentPeriodStart: now,
			CurrentPeriodEnd:  farFuture,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := db.WithContext(ctx).Create(sub).Error; err != nil {
			log.Printf("⚠️  Failed to create subscription for user %s: %v", user.ID, err)
			continue
		}
		log.Printf("✅ Created free subscription for employer: %s (%s)", user.Email, user.ID)
	}

	log.Println("✅ Backfill complete")
	return nil
}
