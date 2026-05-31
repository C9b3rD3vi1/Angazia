package parser

import (
	"context"
	"mime/multipart"
)

// ResumeParser defines the interface for resume parsing
type ResumeParser interface {
	Parse(ctx context.Context, file multipart.File, filename string) (*ParsedResume, error)
	ExtractSkills(ctx context.Context, text string) ([]string, error)
	ExtractExperience(ctx context.Context, text string) (*ExperienceInfo, error)
	ExtractEducation(ctx context.Context, text string) ([]Education, error)
	ExtractContactInfo(ctx context.Context, text string) (*ContactInfo, error)
}

// ParsedResume contains extracted information
type ParsedResume struct {
	FullName        string        `json:"full_name"`
	Email           string        `json:"email"`
	Phone           string        `json:"phone"`
	Location        string        `json:"location"`
	LinkedInURL     string        `json:"linkedin_url"`
	PortfolioURL    string        `json:"portfolio_url"`
	Skills          []string      `json:"skills"`
	Experience      []Experience  `json:"experience"`
	Education       []Education   `json:"education"`
	Summary         string        `json:"summary"`
	TotalExperience int           `json:"total_experience_years"`
	Languages       []string      `json:"languages"`
	Certifications  []string      `json:"certifications"`
}

// Experience represents work experience
type Experience struct {
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Location    string   `json:"location"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	Current     bool     `json:"current"`
	Description []string `json:"description"`
	Skills      []string `json:"skills"`
}

// Education represents educational background
type Education struct {
	Degree      string `json:"degree"`
	Field       string `json:"field"`
	Institution string `json:"institution"`
	Location    string `json:"location"`
	StartYear   int    `json:"start_year"`
	EndYear     int    `json:"end_year"`
	Grade       string `json:"grade"`
}

// ExperienceInfo contains calculated experience metrics
type ExperienceInfo struct {
	TotalYears      int      `json:"total_years"`
	SeniorityLevel  string   `json:"seniority_level"` // entry, junior, mid, senior, lead
	Roles           []string `json:"roles"`
	Industries      []string `json:"industries"`
	TopCompanies    []string `json:"top_companies"`
	CareerProgression []string `json:"career_progression"`
}

// ContactInfo extracted from resume
type ContactInfo struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	LinkedIn    string `json:"linkedin"`
	GitHub      string `json:"github"`
	Twitter     string `json:"twitter"`
	Website     string `json:"website"`
}