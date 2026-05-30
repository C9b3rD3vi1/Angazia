package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	
	"github.com/Angazia/internal/config"
	"github.com/Angazia/internal/handlers"
	"github.com/Angazia/internal/pkg/database"
	"github.com/Angazia/internal/pkg/utils"
	"github.com/Angazia/internal/repository"
	"github.com/Angazia/internal/routes"
	"github.com/Angazia/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Initialize JWT utilities
	utils.InitJWT(cfg)
	
	// Initialize encryption utilities
	utils.InitEncryption(cfg)
	
	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	
	db := database.GetDB()
	
	// Enable PostgreSQL extensions
	if err := database.EnableExtensions(); err != nil {
		log.Println("Warning: Could not enable extensions:", err)
	}
	
	// Run auto-migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	
	// Create additional indexes
	if err := database.CreateIndexes(); err != nil {
		log.Println("Warning: Could not create indexes:", err)
	}
	
	// Run manual migrations
	if err := database.RunMigrations(); err != nil {
		log.Println("Warning: Could not run manual migrations:", err)
	}
	
	// Seed data in development
	if cfg.IsDevelopment() {
		if err := database.SeedData(); err != nil {
			log.Println("Warning: Could not seed data:", err)
		}
	}
	
	// ========== INITIALIZE REPOSITORIES ==========
	userRepo := repository.NewUserRepository(db)
	githubRepo := repository.NewGitHubRepository(db)
	jobRepo := repository.NewJobRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	
	// ========== INITIALIZE SERVICES ==========
	// Core services
	encryptionSvc := services.NewEncryptionService(cfg)
	emailSvc := services.NewEmailService(cfg)
	
	// Business services
	authSvc := services.NewAuthService(cfg, userRepo, encryptionSvc, emailSvc)
	githubSvc := services.NewGitHubService(cfg, githubRepo, userRepo, encryptionSvc)
	jobSvc := services.NewJobService(cfg, jobRepo, applicationRepo, userRepo)
	matchingSvc := services.NewMatchingService(cfg, jobRepo, userRepo, githubRepo)
	
	// ========== INITIALIZE HANDLERS ==========
	authHandler := handlers.NewAuthHandler(authSvc)
	githubHandler := handlers.NewGitHubHandler(githubSvc)
	jobHandler := handlers.NewJobHandler(jobSvc, matchingSvc)
	applicationHandler := handlers.NewApplicationHandler(jobSvc)
	employeeHandler := handlers.NewEmployeeHandler(userRepo, matchingSvc)
	employerHandler := handlers.NewEmployerHandler(jobSvc, userRepo)
	webHandler := handlers.NewWebHandler(jobSvc, matchingSvc, userRepo)
	
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               cfg.AppName,
		Prefork:               cfg.IsProduction(),
		ReadTimeout:           cfg.GetReadTimeout(),
		WriteTimeout:          cfg.GetWriteTimeout(),
		IdleTimeout:           120 * time.Second,
		BodyLimit:             10 * 1024 * 1024, // 10 MB
		DisableStartupMessage: false,
		JSONEncoder:           utils.JSONMarshal,
		JSONDecoder:           utils.JSONUnmarshal,
	})
	
	// Middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Africa/Nairobi",
	}))
	
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))
	
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		ExposeHeaders:    "Content-Length, Content-Type",
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	
	// Static files with caching
	app.Static("/static", "./web/static", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 24 * time.Hour,
		MaxAge:        86400,
	})
	
	// Health check endpoint (no auth)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":      "healthy",
			"timestamp":   time.Now().Unix(),
			"environment": cfg.Environment,
			"version":     cfg.Version,
		})
	})
	
	// Readiness check (for k8s)
	app.Get("/ready", func(c *fiber.Ctx) error {
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not ready",
				"reason": "database connection failed",
			})
		}
		
		return c.JSON(fiber.Map{
			"status": "ready",
		})
	})
	
	// Setup all routes with handlers
	routes.SetupRoutes(app, cfg, authHandler, githubHandler, jobHandler, applicationHandler, employeeHandler, employerHandler, webHandler)
	
	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("🌐 Environment: %s", cfg.Environment)
	log.Printf("🔗 URL: http://localhost:%s", cfg.Port)
	log.Printf("📊 GitHub OAuth: %s", boolToString(cfg.GithubClientID != ""))
	log.Printf("🤖 OpenAI: %s", boolToString(cfg.OpenAIAPIKey != ""))
	log.Printf("📧 Email Service: %s", boolToString(cfg.SMTPHost != ""))
	
	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	utils.WaitForShutdown(app)
}

func boolToString(b bool) string {
	if b {
		return "enabled ✅"
	}
	return "disabled ❌"
}