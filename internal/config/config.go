package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port        string
	Environment string
	CORSAllowOrigins string
	Version     string
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	
	// JWT
	JWTSecret     string
	JWTExpiryHours int
	
	// GitHub OAuth
	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
	

	// Email Provider Configuration
	EmailProvider   string // sendgrid, resend, smtp
	SendGridAPIKey  string
	ResendAPIKey    string
	
	// Redis Configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	
	// Security
	RequireEmailVerification bool
	
	// OpenAI
	OpenAIAPIKey string
	
	// SMTP (Email)
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFromName string
	SMTPFromEmail string
	
	// M-Pesa (Future)
	MPESAConsumerKey    string
	MPESAConsumerSecret string
	MPESAShortcode      string
	MPESAPasskey        string
	
	// IntaSend Payment Gateway
	IntaSendAPIKey         string
	IntaSendAPISecret      string
	IntaSendPublishableKey string

	// Admin
	AdminEmail string

	// App Settings
	AppName     string
	AppURL      string
	AppDomain   string
	PageSize    int
	MaxJobPosts int // Free tier limit
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
	
	config := &Config{
		// Server defaults
		Port:        getEnv("PORT", "3000"),
		Environment: getEnv("ENVIRONMENT", "development"),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", ""),
		Version:        getEnv("VERSION", ""),
		
		// Database defaults
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "kenyan_dev_marketplace"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		
		// JWT defaults
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-key-change-this"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		
		// GitHub defaults
		GithubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GithubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GithubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:3000/auth/github/callback"),
		
		// OpenAI
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		
		// SMTP
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		SMTPFromName: getEnv("SMTP_FROM_NAME", ""),
		SMTPFromEmail: getEnv("SMTP_FROM_EMAIL", ""),

		// Email Provider
		EmailProvider:   getEnv("EMAIL_PROVIDER", "smtp"),
		SendGridAPIKey:  getEnv("SENDGRID_API_KEY", ""),
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		
		// M-Pesa
		MPESAConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
		MPESAConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
		MPESAShortcode:      getEnv("MPESA_SHORTCODE", ""),
		MPESAPasskey:        getEnv("MPESA_PASSKEY", ""),
		
		// IntaSend
		IntaSendAPIKey:         getEnv("INTASEND_API_KEY", ""),
		IntaSendAPISecret:      getEnv("INTASEND_API_SECRET", ""),
		IntaSendPublishableKey: getEnv("INTASEND_PUBLISHABLE_KEY", ""),

		// Admin
		AdminEmail: getEnv("ADMIN_EMAIL", ""),

		// App Settings
		AppName:     getEnv("APP_NAME", "Kenyan Dev Marketplace"),
		AppURL:      getEnv("APP_URL", "http://localhost:3000"),
		AppDomain:   getEnv("APP_DOMAIN", "localhost"),
		PageSize:    getEnvAsInt("PAGE_SIZE", 20),
		MaxJobPosts: getEnvAsInt("MAX_JOB_POSTS", 3),
	}
	
	// Validate required fields
	if config.JWTSecret == "your-super-secret-key-change-this" && config.Environment == "production" {
		log.Fatal("ERROR: JWT_SECRET must be changed in production!")
	}
	
	return config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Africa/Nairobi",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}