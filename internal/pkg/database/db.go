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
	
	// Configure GORM logger
	logLevel := logger.Silent
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
	
	// Connect to database
	dsn := cfg.GetDSN()
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Get underlying SQL DB to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	
	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	log.Println("✅ Database connected successfully")
	return nil
}

// AutoMigrate runs automatic migration for all models
func AutoMigrate() error {
	log.Println("🔄 Running database migrations...")
	
	// Migrate in order to handle dependencies
	err := DB.AutoMigrate(
		// User models
		&models.User{},
		&models.EmployeeProfile{},
		&models.EmployerProfile{},
		
		// GitHub models
		&models.GithubProfile{},
		&models.GithubContribution{},
		&models.GithubRepository{},
		&models.GithubSyncLog{},
		
		// Job models
		&models.Job{},
		&models.Application{},
		&models.SavedJob{},
		&models.JobView{},
		
		// Match models
		&models.Match{},
		&models.MatchSettings{},
		&models.MatchFeedback{},
		&models.TalentPool{},
		&models.TalentPoolCandidate{},
		&models.SearchHistory{},
		
		// 2FA models
		&models.TwoFASecret{},
		&models.TwoFAAuditLog{},
	)
	
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	log.Println("✅ Database migrations completed successfully")
	return nil
}

// CreateIndexes creates additional indexes for performance
func CreateIndexes() error {
	log.Println("📊 Creating additional indexes...")
	
	// User indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		log.Printf("Warning: Could not create users email index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)").Error; err != nil {
		log.Printf("Warning: Could not create users role index: %v", err)
	}
	
	// Employee indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_employee_skills ON employee_profiles USING GIN(skills)").Error; err != nil {
		log.Printf("Warning: Could not create employee skills GIN index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_employee_github ON employee_profiles(github_username)").Error; err != nil {
		log.Printf("Warning: Could not create employee github index: %v", err)
	}
	
	// Job indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_employer ON jobs(employer_id)").Error; err != nil {
		log.Printf("Warning: Could not create jobs employer index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_active ON jobs(is_active, posted_at)").Error; err != nil {
		log.Printf("Warning: Could not create jobs active index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_skills ON jobs USING GIN(required_skills)").Error; err != nil {
		log.Printf("Warning: Could not create jobs skills GIN index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_jobs_location ON jobs(location, is_remote)").Error; err != nil {
		log.Printf("Warning: Could not create jobs location index: %v", err)
	}
	
	// Application indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_job ON applications(job_id)").Error; err != nil {
		log.Printf("Warning: Could not create applications job index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_employee ON applications(employee_id)").Error; err != nil {
		log.Printf("Warning: Could not create applications employee index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status, applied_at)").Error; err != nil {
		log.Printf("Warning: Could not create applications status index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_applications_score ON applications(match_score)").Error; err != nil {
		log.Printf("Warning: Could not create applications score index: %v", err)
	}
	
	// Match indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_matches_employee_score ON matches(employee_id, overall_score)").Error; err != nil {
		log.Printf("Warning: Could not create matches employee score index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_matches_job ON matches(job_id)").Error; err != nil {
		log.Printf("Warning: Could not create matches job index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_matches_expires ON matches(expires_at)").Error; err != nil {
		log.Printf("Warning: Could not create matches expires index: %v", err)
	}
	
	// GitHub indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_github_employee ON github_profiles(employee_id)").Error; err != nil {
		log.Printf("Warning: Could not create github employee index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_github_username ON github_profiles(github_username)").Error; err != nil {
		log.Printf("Warning: Could not create github username index: %v", err)
	}
	
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_github_scores ON github_profiles(overall_score, activity_score)").Error; err != nil {
		log.Printf("Warning: Could not create github scores index: %v", err)
	}
	
	// Talent pool indexes
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_talent_pool_employer ON talent_pools(employer_id)").Error; err != nil {
		log.Printf("Warning: Could not create talent pool employer index: %v", err)
	}
	
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