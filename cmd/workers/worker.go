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
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	
	db := database.GetDB()
	
	// Initialize repositories
	alertRepo := repository.NewAlertRepository(db)
	jobRepo := repository.NewJobRepository(db)
	
	// Initialize services
	emailSvc := services.NewEmailService(cfg, nil, nil) // Note: unsubscribeRepo and tokenService would be injected
	alertSvc := services.NewAlertService(cfg, alertRepo, jobRepo, emailSvc)
	
	// Create cron scheduler
	c := cron.New(cron.WithLocation(time.Local))
	
	// Run daily at 8:00 AM
	if _, err := c.AddFunc("0 8 * * *", func() {
		log.Println("Running daily job alerts...")
		if err := alertSvc.ProcessAllAlerts(context.Background()); err != nil {
			log.Printf("Error processing alerts: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to schedule daily job:", err)
	}
	
	// Run weekly on Monday at 8:00 AM
	if _, err := c.AddFunc("0 8 * * 1", func() {
		log.Println("Running weekly job alerts...")
		if err := alertSvc.ProcessAllAlerts(context.Background()); err != nil {
			log.Printf("Error processing alerts: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to schedule weekly job:", err)
	}
	
	// Run every hour for instant alerts
	if _, err := c.AddFunc("0 * * * *", func() {
		log.Println("Running instant job alerts...")
		if err := alertSvc.ProcessAllAlerts(context.Background()); err != nil {
			log.Printf("Error processing alerts: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to schedule hourly job:", err)
	}
	
	log.Println("Job alert worker started")
	c.Run()
}