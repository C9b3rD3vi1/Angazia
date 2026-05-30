package database

import (
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	
	"github.com/kenyan-dev-marketplace/internal/models"
)

// SeedData seeds the database with initial test data
func SeedData() error {
	log.Println("🌱 Seeding database with initial data...")
	
	// Check if already seeded
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Database already has data, skipping seed")
		return nil
	}
	
	// Create test users
	users := []models.User{
		{
			ID:       uuid.New().String(),
			Email:    "john.doe@example.com",
			Role:     models.RoleEmployee,
			IsVerified: true,
			IsActive: true,
		},
		{
			ID:       uuid.New().String(),
			Email:    "jane.smith@example.com",
			Role:     models.RoleEmployee,
			IsVerified: true,
			IsActive: true,
		},
		{
			ID:       uuid.New().String(),
			Email:    "safaricom@example.com",
			Role:     models.RoleEmployer,
			IsVerified: true,
			IsActive: true,
		},
		{
			ID:       uuid.New().String(),
			Email:    "andela@example.com",
			Role:     models.RoleEmployer,
			IsVerified: true,
			IsActive: true,
		},
	}
	
	// Hash passwords (default: "password123")
	for i := range users {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		users[i].PasswordHash = string(hashedPassword)
	}
	
	if err := DB.Create(&users).Error; err != nil {
		return err
	}
	
	// Create employee profiles
	employeeProfiles := []models.EmployeeProfile{
		{
			UserID:            users[0].ID,
			FullName:          "John Doe",
			Headline:          "Senior Full Stack Developer",
			Bio:               "Experienced developer with 5+ years in React and Go. Passionate about building scalable applications.",
			Location:          "Nairobi, Kenya",
			IsRemoteOnly:      false,
			ExperienceLevel:   "senior",
			YearsOfExperience: 5,
			Skills:            []string{"Go", "React", "TypeScript", "PostgreSQL", "Docker", "Kubernetes"},
			IsVisible:         true,
			IsAvailable:       true,
			GithubConnected:   true,
			GithubUsername:    "johndoe",
		},
		{
			UserID:            users[1].ID,
			FullName:          "Jane Smith",
			Headline:          "DevOps Engineer",
			Bio:               "Cloud infrastructure specialist with AWS and Kubernetes expertise.",
			Location:          "Mombasa, Kenya",
			IsRemoteOnly:      true,
			ExperienceLevel:   "mid",
			YearsOfExperience: 3,
			Skills:            []string{"AWS", "Kubernetes", "Terraform", "Python", "CI/CD", "Linux"},
			IsVisible:         true,
			IsAvailable:       true,
			GithubConnected:   true,
			GithubUsername:    "janesmith",
		},
	}
	
	if err := DB.Create(&employeeProfiles).Error; err != nil {
		return err
	}
	
	// Create employer profiles
	employerProfiles := []models.EmployerProfile{
		{
			UserID:             users[2].ID,
			CompanyName:        "Safaricom",
			CompanyDescription: "Leading telecommunications company in Kenya",
			Industry:           "Telecommunications",
			CompanySize:        "500+",
			Location:           "Nairobi, Kenya",
			VerificationStatus: "verified",
		},
		{
			UserID:             users[3].ID,
			CompanyName:        "Andela",
			CompanyDescription: "Global talent network connecting developers with top companies",
			Industry:           "Technology",
			CompanySize:        "500+",
			Location:           "Nairobi, Kenya",
			VerificationStatus: "verified",
		},
	}
	
	if err := DB.Create(&employerProfiles).Error; err != nil {
		return err
	}
	
	// Create jobs
	jobs := []models.Job{
		{
			ID:               uuid.New().String(),
			EmployerID:       users[2].ID,
			Title:            "Senior Go Developer",
			Description:      "Looking for an experienced Go developer to build microservices for our mobile payment platform.",
			Requirements:     "5+ years experience with Go, experience with microservices, knowledge of PostgreSQL",
			RequiredSkills:   []string{"Go", "Microservices", "PostgreSQL", "REST APIs", "Git"},
			NiceToHaveSkills: []string{"Kubernetes", "gRPC", "Redis"},
			ExperienceLevel:  "senior",
			MinExperience:    5,
			MaxExperience:    8,
			SalaryMin:        150000,
			SalaryMax:        250000,
			SalaryCurrency:   "KES",
			IsSalaryVisible:  true,
			Location:         "Nairobi, Kenya",
			IsRemote:         false,
			IsHybrid:         true,
			EmploymentType:   "full-time",
			IsActive:         true,
			IsFeatured:       true,
			ExpiresAt:        timePtr(time.Now().AddDate(0, 1, 0)),
		},
		{
			ID:               uuid.New().String(),
			EmployerID:       users[2].ID,
			Title:            "React Frontend Developer",
			Description:      "Join our team to build responsive web applications for millions of users.",
			Requirements:     "3+ years with React, TypeScript, and modern CSS frameworks",
			RequiredSkills:   []string{"React", "TypeScript", "TailwindCSS", "Redux", "JavaScript"},
			NiceToHaveSkills: []string{"Next.js", "Jest", "GraphQL"},
			ExperienceLevel:  "mid",
			MinExperience:    3,
			MaxExperience:    5,
			SalaryMin:        100000,
			SalaryMax:        180000,
			SalaryCurrency:   "KES",
			IsSalaryVisible:  true,
			Location:         "Nairobi, Kenya",
			IsRemote:         true,
			IsHybrid:         false,
			EmploymentType:   "full-time",
			IsActive:         true,
			IsFeatured:       true,
			ExpiresAt:        timePtr(time.Now().AddDate(0, 1, 0)),
		},
		{
			ID:               uuid.New().String(),
			EmployerID:       users[3].ID,
			Title:            "DevOps Engineer",
			Description:      "Looking for a DevOps engineer to manage our cloud infrastructure.",
			Requirements:     "Experience with AWS, Kubernetes, and CI/CD pipelines",
			RequiredSkills:   []string{"AWS", "Kubernetes", "Docker", "Jenkins", "Terraform"},
			NiceToHaveSkills: []string{"Prometheus", "Grafana", "ELK Stack"},
			ExperienceLevel:  "mid",
			MinExperience:    3,
			MaxExperience:    6,
			SalaryMin:        120000,
			SalaryMax:        200000,
			SalaryCurrency:   "KES",
			IsSalaryVisible:  true,
			Location:         "Remote",
			IsRemote:         true,
			IsHybrid:         false,
			EmploymentType:   "full-time",
			IsActive:         true,
			ExpiresAt:        timePtr(time.Now().AddDate(0, 1, 0)),
		},
	}
	
	if err := DB.Create(&jobs).Error; err != nil {
		return err
	}
	
	// Create match settings for employees
	matchSettings := []models.MatchSettings{
		{
			EmployeeID:    users[0].ID,
			SkillsWeight:  40,
			ExperienceWeight: 30,
			LocationWeight: 15,
			SalaryWeight:   10,
			CultureWeight:  5,
			MinSalary:      120000,
			MaxSalary:      300000,
			IsOpenToRemote: true,
		},
		{
			EmployeeID:    users[1].ID,
			SkillsWeight:  35,
			ExperienceWeight: 25,
			LocationWeight: 10,
			SalaryWeight:   20,
			CultureWeight:  10,
			MinSalary:      80000,
			MaxSalary:      200000,
			IsOpenToRemote: true,
		},
	}
	
	if err := DB.Create(&matchSettings).Error; err != nil {
		return err
	}
	
	// Create matches
	matches := []models.Match{
		{
			ID:              uuid.New().String(),
			JobID:           jobs[0].ID,
			EmployeeID:      users[0].ID,
			OverallScore:    85,
			SkillsScore:     90,
			ExperienceScore: 85,
			LocationScore:   80,
			SalaryScore:     85,
			MatchingSkills:  []string{"Go", "PostgreSQL", "Microservices"},
			MissingSkills:   []string{"Kubernetes"},
			MatchReason:     "Strong match with required skills and experience level",
			ExpiresAt:       time.Now().AddDate(0, 1, 0),
		},
		{
			ID:              uuid.New().String(),
			JobID:           jobs[2].ID,
			EmployeeID:      users[1].ID,
			OverallScore:    92,
			SkillsScore:     95,
			ExperienceScore: 90,
			LocationScore:   100,
			SalaryScore:     85,
			MatchingSkills:  []string{"AWS", "Kubernetes", "Docker", "Terraform"},
			MissingSkills:   []string{},
			MatchReason:     "Excellent match! Your skills align perfectly with our requirements",
			ExpiresAt:       time.Now().AddDate(0, 1, 0),
		},
	}
	
	if err := DB.Create(&matches).Error; err != nil {
		return err
	}
	
	// Create GitHub profiles
	githubProfiles := []models.GithubProfile{
		{
			EmployeeID:        users[0].ID,
			GithubUsername:    "johndoe",
			PublicRepos:       45,
			Followers:         128,
			Following:         45,
			TotalCommits:      1240,
			ContributionStreak: 12,
			OverallScore:      78,
			ActivityScore:     82,
			QualityScore:      75,
			LastSyncedAt:      time.Now(),
		},
		{
			EmployeeID:        users[1].ID,
			GithubUsername:    "janesmith",
			PublicRepos:       28,
			Followers:         89,
			Following:         32,
			TotalCommits:      890,
			ContributionStreak: 8,
			OverallScore:      71,
			ActivityScore:     75,
			QualityScore:      68,
			LastSyncedAt:      time.Now(),
		},
	}
	
	if err := DB.Create(&githubProfiles).Error; err != nil {
		log.Printf("Warning: Could not seed GitHub profiles: %v", err)
	}
	
	log.Printf("✅ Seeded %d users, %d jobs, %d matches", len(users), len(jobs), len(matches))
	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// ResetDatabase truncates all tables (for testing only)
func ResetDatabase() error {
	log.Println("⚠️  Resetting database...")
	
	tables := []string{
		"match_feedback", "match_settings", "matches",
		"applications", "job_views", "saved_jobs", "jobs",
		"github_sync_logs", "github_repositories", "github_contributions", "github_profiles",
		"employee_profiles", "employer_profiles",
		"users",
	}
	
	for _, table := range tables {
		if err := DB.Exec("TRUNCATE TABLE " + table + " CASCADE").Error; err != nil {
			log.Printf("Warning: Could not truncate %s: %v", table, err)
		}
	}
	
	return SeedData()
}