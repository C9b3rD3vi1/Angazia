package main

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/database"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := database.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	db := database.GetDB()

	// Repositories
	alertRepo := repository.NewAlertRepository(db)
	jobRepo := repository.NewJobRepository(db)
	userRepo := repository.NewUserRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	unsubscribeRepo := repository.NewUnsubscribeRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// Services
	tokenService := services.NewTokenService(cfg)
	emailSvc := services.NewEmailService(cfg, unsubscribeRepo, tokenService)
	notificationSvc := services.NewNotificationService(cfg, notificationRepo, userRepo, emailSvc)
	alertSvc := services.NewAlertService(cfg, alertRepo, jobRepo, emailSvc, notificationSvc)
	billingSvc := services.NewSubscriptionService(cfg, subscriptionRepo, paymentRepo, userRepo, jobRepo)

	ctx := context.Background()
	c := cron.New(cron.WithLocation(time.Local))

	// Daily at 8:00 AM — job alerts + subscription expiry
	if _, err := c.AddFunc("0 8 * * *", func() {
		log.Println("[Worker] Running daily tasks...")

		log.Println("[Worker] Processing job alerts...")
		if err := alertSvc.ProcessAllAlerts(ctx); err != nil {
			log.Printf("[Worker] Alert processing error: %v", err)
		}

		log.Println("[Worker] Expiring overdue subscriptions...")
		if err := billingSvc.ExpireSubscriptions(ctx); err != nil {
			log.Printf("[Worker] Expiry error: %v", err)
		}

		log.Println("[Worker] Processing payment retries...")
		if err := billingSvc.ProcessPaymentRetries(ctx); err != nil {
			log.Printf("[Worker] Payment retry error: %v", err)
		}

		log.Println("[Worker] Deleting expired payment intents...")
		paymentRepo.DeleteExpiredIntents(ctx)

		log.Println("[Worker] Resetting expired usage records...")
		subscriptionRepo.ResetExpiredUsage(ctx)

		log.Println("[Worker] Daily tasks complete.")
	}); err != nil {
		log.Fatal("Failed to schedule daily job:", err)
	}

	// Weekly on Monday at 8:00 AM — job alerts for weekly frequency
	if _, err := c.AddFunc("0 8 * * 1", func() {
		log.Println("[Worker] Running weekly job alerts...")
		if err := alertSvc.ProcessAllAlerts(ctx); err != nil {
			log.Printf("[Worker] Weekly alert error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to schedule weekly job:", err)
	}

	// Every hour — instant alerts + expiry scanning
	if _, err := c.AddFunc("0 * * * *", func() {
		log.Println("[Worker] Running hourly tasks...")

		if err := alertSvc.ProcessAllAlerts(ctx); err != nil {
			log.Printf("[Worker] Hourly alert error: %v", err)
		}

		if err := billingSvc.ExpireSubscriptions(ctx); err != nil {
			log.Printf("[Worker] Hourly expiry error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to schedule hourly job:", err)
	}

	log.Println("[Worker] Billing and alert worker started")
	c.Run()
}
