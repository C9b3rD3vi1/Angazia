package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	
	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/handlers"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/ai"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/database"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/elasticsearch"
	"github.com/C9b3rD3vi1/Angazia/internal/middleware"
	pkgredis "github.com/C9b3rD3vi1/Angazia/internal/pkg/redis"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
	"github.com/C9b3rD3vi1/Angazia/internal/routes"
	"github.com/C9b3rD3vi1/Angazia/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	// Initialize utilities
	utils.InitJWT(cfg)
	utils.InitEncryption(cfg)
	
	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	
	db := database.GetDB()
	
	// Initialize Redis for token blacklisting
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	
	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Token blacklisting will use database only.", err)
		redisClient = nil
	}
	
	// Initialize wrapped Redis client (sessions, queues, cache)
	redisWrapped, err := pkgredis.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Warning: Wrapped Redis client failed: %v. Some features will be degraded.", err)
		redisWrapped = nil
	}
	
	// Initialize Elasticsearch
	esClient, err := elasticsearch.NewESClient(cfg)
	if err != nil {
		log.Printf("Warning: Elasticsearch not available: %v. Search will use database only.", err)
		esClient = nil
	}
	
	if esClient != nil {
		ctx := context.Background()
		if err := esClient.CreateIndex(ctx, "jobs", elasticsearch.JobIndexMapping); err != nil {
			log.Printf("Warning: Could not create jobs index: %v", err)
		}
		if err := esClient.CreateIndex(ctx, "candidates", elasticsearch.CandidateIndexMapping); err != nil {
			log.Printf("Warning: Could not create candidates index: %v", err)
		}
		if err := esClient.CreateIndex(ctx, "companies", elasticsearch.CompanyIndexMapping); err != nil {
			log.Printf("Warning: Could not create companies index: %v", err)
		}
	}
	
	// Run database migrations
	if err := initializeDatabase(cfg, db); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	githubRepo := repository.NewGitHubRepository(db)
	unsubscribeRepo := repository.NewUnsubscribeRepository(db)
	// Initialize job repository and service
	jobRepo := repository.NewJobRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	searchRepo := repository.NewSearchRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	talentRepo := repository.NewTalentPoolRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	candidateAnalyticsRepo := repository.NewCandidateAnalyticsRepository(db)
	
	// Initialize services
	tokenService := services.NewTokenService(cfg)
	jobSvc := services.NewJobService(cfg, jobRepo, userRepo)
	emailSvc := services.NewEmailService(cfg, unsubscribeRepo, tokenService)
	authSvc := services.NewAuthService(cfg, userRepo, db, emailSvc, redisClient)
	githubSvc := services.NewGitHubService(cfg, githubRepo, userRepo)
	companyService := services.NewCompanyService(cfg, companyRepo, userRepo, jobRepo, emailSvc)
	
	// AI Provider
	aiFactory := ai.NewProviderFactory(ai.DefaultConfig())
	aiProvider, err := aiFactory.GetProvider()
	if err != nil {
		log.Printf("Warning: AI provider not available: %v. AI-powered matching features will be unavailable.", err)
	}
	
	// Additional services
	applicationSvc := services.NewApplicationService(cfg, applicationRepo, jobRepo, userRepo, emailSvc)
	notificationSvc := services.NewNotificationService(cfg, notificationRepo, emailSvc)
	matchingSvc := services.NewMatchingService(cfg, aiProvider, jobRepo, userRepo, githubRepo, matchRepo)
	alertSvc := services.NewAlertService(cfg, alertRepo, jobRepo, emailSvc)
	searchSvc := services.NewSearchService(cfg, searchRepo, jobRepo, userRepo)
	adminSvc := services.NewAdminService(cfg, adminRepo)
	analyticsSvc := services.NewAnalyticsService(cfg, analyticsRepo)
	candidateAnalyticsSvc := services.NewCandidateAnalyticsService(cfg, candidateAnalyticsRepo, jobRepo, applicationRepo, matchingSvc)
	talentPoolSvc := services.NewTalentPoolService(cfg, talentRepo, userRepo, jobRepo, matchingSvc)
	subscriptionSvc := services.NewSubscriptionService(cfg, subscriptionRepo, paymentRepo, userRepo, jobRepo)
	profileSvc := services.NewProfileService(cfg, userRepo, githubRepo)
	
	// Initialize ES-based services
	var (
		syncService *elasticsearch.SyncService
	)
	
	if esClient != nil {
		jobIndexer := elasticsearch.NewJobIndexer(esClient, jobRepo, userRepo)
		candidateIndexer := elasticsearch.NewCandidateIndexer(esClient, userRepo)
		companyIndexer := elasticsearch.NewCompanyIndexer(esClient, userRepo)
		
		syncService = elasticsearch.NewSyncService(esClient, jobIndexer, candidateIndexer, companyIndexer)
		syncService.Start(context.Background())
		
		_ = services.NewSearchESService(esClient, jobRepo, userRepo)
	}
	
	if redisWrapped != nil {
		_ = services.NewCacheService(redisWrapped, jobRepo, userRepo)
		_ = pkgredis.NewSessionManager(redisWrapped, 24*time.Hour)
	}
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	// Initialize job handler
	jobHandler := handlers.NewJobHandler(jobSvc)

	reviewHandler := handlers.NewReviewHandler(companyService)
	githubHandler := handlers.NewGitHubHandler(githubSvc)
	unsubscribeHandler := handlers.NewUnsubscribeHandler(unsubscribeRepo, emailSvc)
	applicationHandler := handlers.NewApplicationHandler(applicationSvc)
	matchingHandler := handlers.NewMatchingHandler(matchingSvc)
	searchHandler := handlers.NewSearchHandler(searchSvc)
	adminHandler := handlers.NewAdminHandler(adminSvc)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsSvc)
	candidateAnalyticsHandler := handlers.NewCandidateAnalyticsHandler(candidateAnalyticsSvc)
	talentPoolHandler := handlers.NewTalentPoolHandler(talentPoolSvc)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionSvc)
	notificationHandler := handlers.NewNotificationHandler(notificationSvc)
	alertHandler := handlers.NewAlertHandler(alertSvc)
	planHandler := handlers.NewAdminPlanHandler(subscriptionSvc)
	webHandler := handlers.NewWebHandler(jobSvc, companyService)
	companyHandler := handlers.NewCompanyHandler(companyService)
	resumeHandler := handlers.NewResumeHandler(profileSvc)
	websocketHandler := handlers.NewWebSocketHandler()
	
	// Create Fiber app with Go html/template engine
	engine := html.New(cfg.TemplateDir, ".html")
	engine.Reload(cfg.IsDevelopment())
	engine.AddFunc("unescape", func(s string) template.HTML {
		return template.HTML(s)
	})
	engine.AddFunc("formatNumber", func(n interface{}) string {
		var i int64
		switch v := n.(type) {
		case int:
			i = int64(v)
		case int64:
			i = v
		case float64:
			i = int64(v)
		default:
			return fmt.Sprintf("%v", v)
		}
		s := strconv.FormatInt(i, 10)
		if len(s) <= 3 {
			return s
		}
		var result []byte
		for j, c := range s {
			if j > 0 && (len(s)-j)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		return string(result)
	})
	engine.AddFunc("iterate", func(count int) []int {
		r := make([]int, count)
		for i := 0; i < count; i++ {
			r[i] = i
		}
		return r
	})

	app := fiber.New(fiber.Config{
		AppName:               cfg.AppName,
		Prefork:               cfg.IsProduction(),
		Views:                 engine,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		BodyLimit:             10 * 1024 * 1024,
		DisableStartupMessage: false,
	})
	
	// Setup middleware
	setupMiddleware(app, cfg)
	
	// Static files
	app.Static("/static", cfg.StaticDir, fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		CacheDuration: 24 * time.Hour,
		MaxAge:        86400,
	})
	
	// Health check endpoints
	setupHealthEndpoints(app, cfg, db)
	
	// API routes
	api := app.Group("/api/v1")
	
	// Auth routes
	routes.SetupAuthRoutes(api, authHandler)
	
	// GitHub routes
	routes.SetupGitHubRoutes(api, githubHandler)
	
	// Unsubscribe routes
	routes.SetupUnsubscribeRoutes(api, unsubscribeHandler)

	// Job routes
	routes.SetupJobRoutes(api, jobHandler)

	// Setup review routes
	routes.SetupReviewRoutes(api, reviewHandler)
	
	// Application routes
	routes.SetupApplicationRoutes(api, applicationHandler)
	
	// Matching routes
	routes.SetupMatchingRoutes(api, matchingHandler)
	
	// Search routes
	routes.SetupSearchRoutes(api, searchHandler)
	
	// Notification routes
	routes.SetupNotificationRoutes(api, notificationHandler)
	
	// Alert routes
	routes.SetupAlertRoutes(api, alertHandler)
	
	// Company routes
	routes.SetupCompanyRoutes(api, companyHandler)
	
	// Resume routes
	routes.SetupResumeRoutes(api, resumeHandler)
	
	// Analytics routes
	routes.SetupAnalyticsRoutes(api, analyticsHandler)
	
	// Candidate analytics routes
	routes.SetupCandidateAnalyticsRoutes(api, candidateAnalyticsHandler)
	
	// Talent pool routes
	routes.SetupTalentPoolRoutes(api, talentPoolHandler)
	
	// Admin routes
	routes.SetupAdminRoutes(api, adminHandler)
	
	// Subscription routes
	routes.SetupSubscriptionRoutes(api, subscriptionHandler)
	
	// Plan routes (public + admin)
	routes.SetupPublicPlanRoutes(api, planHandler)
	routes.SetupAdminPlanRoutes(api, planHandler)
	
	// 2FA routes
	twoFARepo := repository.NewTwoFARepository(db)
	smsProviderFactory := services.NewSMSProviderFactory(cfg)
	smsProvider := smsProviderFactory.GetProvider()
	twoFAService := services.NewTwoFAService(cfg, twoFARepo, userRepo, smsProvider, emailSvc, redisClient)
	twoFAHandler := handlers.NewTwoFAHandler(twoFAService)
	routes.SetupTwoFARoutes(api, twoFAHandler)
	routes.SetupTwoFAGlobalMiddleware(app, twoFAService)
	
	// Web routes (HTML pages)
	routes.SetupWebRoutes(app, webHandler, authHandler, authSvc)
	
	// WebSocket routes
	routes.SetupWebSocketRoutes(app, websocketHandler)
	
	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("🌐 Environment: %s", cfg.Environment)
	log.Printf("🔗 URL: http://localhost:%s", cfg.Port)
	log.Printf("📧 Email Provider: %s", cfg.EmailProvider)
	log.Printf("📊 GitHub OAuth: %s", boolToString(cfg.GithubClientID != ""))
	
	// Start server with graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	utils.WaitForShutdown(app)
	
	// Cleanup Elasticsearch sync service
	if syncService != nil {
		syncService.Stop()
	}
	
	// Close wrapped Redis client
	if redisWrapped != nil {
		redisWrapped.Close()
	}
}

func initializeDatabase(cfg *config.Config, db *gorm.DB) error {
	if err := database.EnableExtensions(); err != nil {
		log.Println("Warning: Could not enable extensions:", err)
	}
	
	if err := database.AutoMigrate(); err != nil {
		return err
	}
	
	if err := database.CreateIndexes(); err != nil {
		log.Println("Warning: Could not create indexes:", err)
	}
	
	if err := database.RunMigrations(); err != nil {
		log.Println("Warning: Could not run manual migrations:", err)
	}
	
	if cfg.IsDevelopment() {
		if err := database.SeedData(); err != nil {
			log.Println("Warning: Could not seed data:", err)
		}
	}
	
	return nil
}

func setupMiddleware(app *fiber.App, cfg *config.Config) {
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Africa/Nairobi",
	}))

	rateLimitConfig := middleware.DefaultRateLimitConfig()
	app.Use(middleware.RateLimit(rateLimitConfig))
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))
	
	allowOrigins := cfg.CORSAllowOrigins
	if allowOrigins == "" {
		allowOrigins = "*"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Client-Type",
		ExposeHeaders:    "Content-Length, Content-Type",
		AllowCredentials: allowOrigins != "*",
		MaxAge:           86400,
	}))
}

func setupHealthEndpoints(app *fiber.App, cfg *config.Config, db *gorm.DB) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return utils.Success(c, fiber.Map{
			"status":      "healthy",
			"timestamp":   time.Now().Unix(),
			"environment": cfg.Environment,
			"version":     cfg.Version,
		})
	})
	
	app.Get("/ready", func(c *fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return utils.Error(c, fiber.StatusServiceUnavailable, "database connection failed")
		}
		
		return utils.Success(c, fiber.Map{
			"status": "ready",
		})
	})
}

func boolToString(b bool) string {
	if b {
		return "enabled ✅"
	}
	return "disabled ❌"
}