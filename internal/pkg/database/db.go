package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB(cfg *config.Config) error {
	var err error

	logLevel := logger.Silent
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	dsn := cfg.GetDSN()
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("✅ Database connected successfully")
	return nil
}

// AutoMigrate runs automatic migration for all models in correct order
func AutoMigrate() error {
	log.Println("🔄 Running database migrations...")

	// Define migration order (tables without dependencies first)
	modelsToMigrate := []interface{}{
		// Level 1: Core tables (no dependencies)
		&models.User{},

		// Level 2: Profile tables (depend on users)
		&models.EmployeeProfile{},
		&models.EmployerProfile{},

		// Level 3: GitHub models
		&models.GithubProfile{},
		&models.GithubContribution{},
		&models.GithubRepository{},
		&models.GithubSyncLog{},

		// Level 4: Job related (depend on employer profiles)
		&models.Job{},
		&models.SavedJob{},
		&models.JobView{},
		&models.Application{},

		// Level 5: Match and Talent Pool
		&models.Match{},
		&models.MatchSettings{},
		&models.MatchFeedback{},
		&models.TalentPool{},
		&models.TalentPoolCandidate{},
		&models.SearchHistory{},

		// Level 6: Payment (no foreign key constraints)
		&models.Payment{},
		&models.PaymentIntent{},
		&models.PaymentMethod{},

		// Level 7: Subscription (no foreign key constraints)
		&models.Subscription{},
		&models.SubscriptionPlan{},
		&models.SubscriptionHistory{},
		&models.SubscriptionPlanFeature{},
		&models.SubscriptionUsage{},

		// Level 8: Invoice (no foreign key constraints)
		&models.Invoice{},
		&models.InvoiceItem{},

		// Level 9: 2FA models
		&models.TwoFASecret{},
		//	&models.TrustedDevice{},
		&models.TwoFAAuditLog{},

		// Level 9b: Session models
		&models.UserSession{},

		// Level 10: Admin models
		&models.AdminActionLog{},
		&models.ModerationQueue{},
		&models.SystemSetting{},
		&models.ReportReason{},

		// Level 11: Notification models
		&models.Notification{},
		&models.NotificationPreferences{},

		// Level 12: Company models
		&models.CompanyVerification{},
		&models.TrustBadge{},
		&models.CompanyReview{},
		&models.CompanyAnalytics{},
		&models.TeamInvitation{},

		// Level 13: Alert/SavedSearch models
		&models.AlertSettings{},
		&models.SavedSearch{},
		&models.AlertHistory{},

		// Level 14: Messaging models
		&models.Conversation{},
		&models.ConversationParticipant{},
		&models.Message{},

		// Level 15: Testimonial and Contact
		&models.Testimonial{},
		&models.ContactSubmission{},

		// Level 16: GDPR models
	}

	// Migrate each model individually to isolate errors
	for _, model := range modelsToMigrate {
		if err := DB.AutoMigrate(model); err != nil {
			log.Printf("Warning: Could not migrate %T: %v", model, err)
			// Continue with other migrations
		}
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// CreateIndexes creates additional indexes for performance
func CreateIndexes() error {
	log.Println("📊 Creating additional indexes...")

	// User indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)")

	// Employee indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_employee_skills ON employee_profiles USING GIN(skills)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_employee_github ON employee_profiles(github_username)")

	// Job indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_employer ON jobs(employer_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_active ON jobs(is_active, posted_at)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_skills ON jobs USING GIN(required_skills)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_location ON jobs(location, is_remote)")

	// Application indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_job ON applications(job_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_employee ON applications(employee_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status, applied_at)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_score ON applications(match_score)")

	// Match indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_matches_employee_score ON matches(employee_id, overall_score)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_matches_job ON matches(job_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_matches_expires ON matches(expires_at)")

	// GitHub indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_github_employee ON github_profiles(employee_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_github_username ON github_profiles(github_username)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_github_scores ON github_profiles(overall_score, activity_score)")

	// Talent pool indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_talent_pool_employer ON talent_pools(employer_id)")

	// Payment indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference)")

	// Subscription indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date)")

	// Invoice indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON invoices(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_invoices_invoice_number ON invoices(invoice_number)")

	log.Println("✅ Additional indexes created successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Transaction runs a function within a database transaction
func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}

// EnableExtensions enables PostgreSQL extensions
func EnableExtensions() error {
	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
		"CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"",
	}

	for _, ext := range extensions {
		if err := DB.Exec(ext).Error; err != nil {
			log.Printf("Warning: Could not enable extension: %v", err)
		}
	}

	log.Println("✅ PostgreSQL extensions enabled")
	return nil
}
