package ai

import (
	"strings"
)

// InsightsGenerator generates career insights and recommendations
type InsightsGenerator struct {
	matcher *Matcher
}

func NewInsightsGenerator() *InsightsGenerator {
	return &InsightsGenerator{
		matcher: NewMatcher(),
	}
}

// CareerInsights contains career-related recommendations
type CareerInsights struct {
	StrengthAreas      []string          `json:"strength_areas"`
	ImprovementAreas   []string          `json:"improvement_areas"`
	RecommendedRoles   []string          `json:"recommended_roles"`
	SalaryRange        SalaryRange       `json:"salary_range"`
	MarketDemand       int               `json:"market_demand"` // 0-100
	GrowthOpportunities []string         `json:"growth_opportunities"`
}

// SalaryRange represents salary expectations
type SalaryRange struct {
	Min int    `json:"min"`
	Max int    `json:"max"`
	Avg int    `json:"avg"`
	Currency string `json:"currency"`
}

// GenerateCareerInsights creates career recommendations based on profile
func (g *InsightsGenerator) GenerateCareerInsights(skills []string, experienceYears int, location string) *CareerInsights {
	insights := &CareerInsights{
		StrengthAreas:      []string{},
		ImprovementAreas:   []string{},
		RecommendedRoles:   []string{},
		GrowthOpportunities: []string{},
		SalaryRange: SalaryRange{
			Min:      0,
			Max:      0,
			Avg:      0,
			Currency: "KES",
		},
	}
	
	// Determine strength areas based on skills
	skillCategories := map[string][]string{
		"backend":  {"go", "python", "java", "nodejs", "rust", "scala", "php", "ruby"},
		"frontend": {"react", "vue", "angular", "javascript", "typescript", "html", "css", "tailwind"},
		"devops":   {"docker", "kubernetes", "aws", "azure", "gcp", "terraform", "jenkins", "cicd"},
		"data":     {"python", "pandas", "numpy", "tensorflow", "pytorch", "sql", "etl", "spark"},
		"mobile":   {"react native", "flutter", "swift", "kotlin", "android", "ios"},
		"security": {"cybersecurity", "penetration testing", "owasp", "burp suite", "nmap"},
	}
	
	categoryStrength := make(map[string]int)
	for category, categorySkills := range skillCategories {
		count := 0
		for _, skill := range skills {
			for _, catSkill := range categorySkills {
				if strings.Contains(strings.ToLower(skill), strings.ToLower(catSkill)) {
					count++
					break
				}
			}
		}
		if count > 0 {
			categoryStrength[category] = count
			insights.StrengthAreas = append(insights.StrengthAreas, category)
		}
	}
	
	// Recommend roles based on skills and experience
	if experienceYears >= 5 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Senior Engineer", "Tech Lead", "Architect")
	} else if experienceYears >= 3 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Mid-Level Engineer", "Team Lead")
	} else if experienceYears >= 1 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Junior Engineer", "Associate Developer")
	} else {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Entry Level Developer", "Intern")
	}
	
	// Add role suggestions based on strengths
	if categoryStrength["backend"] > 0 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Backend Developer")
	}
	if categoryStrength["frontend"] > 0 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Frontend Developer")
	}
	if categoryStrength["devops"] > 0 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "DevOps Engineer", "Site Reliability Engineer")
	}
	if categoryStrength["data"] > 0 {
		insights.RecommendedRoles = append(insights.RecommendedRoles, "Data Scientist", "Data Engineer")
	}
	
	// Calculate market demand based on skills
	demandScore := 50
	inDemandSkills := map[string]int{
		"go": 15, "python": 10, "react": 15, "typescript": 10,
		"aws": 15, "kubernetes": 15, "docker": 10, "terraform": 10,
	}
	for _, skill := range skills {
		if points, ok := inDemandSkills[strings.ToLower(skill)]; ok {
			demandScore += points
		}
	}
	if demandScore > 100 {
		demandScore = 100
	}
	insights.MarketDemand = demandScore
	
	// Calculate salary range based on experience and location
	if location == "Nairobi" || location == "Mombasa" || location == "Kisumu" {
		insights.SalaryRange.Min = 50000 + (experienceYears * 20000)
		insights.SalaryRange.Max = 80000 + (experienceYears * 30000)
	} else {
		insights.SalaryRange.Min = 40000 + (experienceYears * 15000)
		insights.SalaryRange.Max = 60000 + (experienceYears * 20000)
	}
	insights.SalaryRange.Avg = (insights.SalaryRange.Min + insights.SalaryRange.Max) / 2
	
	// Growth opportunities
	insights.GrowthOpportunities = []string{
		"Build a portfolio of side projects",
		"Contribute to open source",
		"Attend tech meetups and conferences",
		"Get certified in cloud platforms",
		"Learn system design and architecture",
	}
	
	return insights
}

// MarketInsights provides market-wide insights
type MarketInsights struct {
	TopSkills       []SkillDemand `json:"top_skills"`
	TrendingRoles   []string      `json:"trending_roles"`
	AverageSalaries map[string]int `json:"average_salaries"`
	HiringTrend     string        `json:"hiring_trend"` // increasing, stable, decreasing
}

type SkillDemand struct {
	Name        string `json:"name"`
	DemandScore int    `json:"demand_score"`
	GrowthRate  int    `json:"growth_rate"` // percentage
}

// GetMarketInsights returns current market insights for Kenyan tech industry
func (g *InsightsGenerator) GetMarketInsights() *MarketInsights {
	return &MarketInsights{
		TopSkills: []SkillDemand{
			{Name: "Go", DemandScore: 95, GrowthRate: 30},
			{Name: "React", DemandScore: 92, GrowthRate: 25},
			{Name: "Python", DemandScore: 88, GrowthRate: 20},
			{Name: "Kubernetes", DemandScore: 85, GrowthRate: 45},
			{Name: "AWS", DemandScore: 82, GrowthRate: 35},
			{Name: "TypeScript", DemandScore: 80, GrowthRate: 40},
			{Name: "Docker", DemandScore: 78, GrowthRate: 25},
			{Name: "PostgreSQL", DemandScore: 75, GrowthRate: 15},
		},
		TrendingRoles: []string{
			"DevOps Engineer",
			"Cloud Architect",
			"AI/ML Engineer",
			"Security Analyst",
			"Full Stack Developer",
		},
		AverageSalaries: map[string]int{
			"Entry Level":    60000,
			"Junior":         90000,
			"Mid-Level":      150000,
			"Senior":         250000,
			"Lead":           350000,
			"Principal":      500000,
		},
		HiringTrend: "increasing",
	}
}

// SkillGapReport generates a detailed skill gap report
type SkillGapReport struct {
	SkillName        string   `json:"skill_name"`
	CurrentLevel     string   `json:"current_level"` // beginner, intermediate, advanced
	RequiredLevel    string   `json:"required_level"`
	Gap              string   `json:"gap"` // none, minor, moderate, severe
	LearningPath     []string `json:"learning_path"`
	EstimatedHours   int      `json:"estimated_hours"`
	Resources        []string `json:"resources"`
}

// GenerateSkillGapReport creates a report for a specific skill
func (g *InsightsGenerator) GenerateSkillGapReport(skillName string, currentLevel, requiredLevel string) *SkillGapReport {
	gap := "none"
	if currentLevel != requiredLevel {
		switch {
		case currentLevel == "beginner" && requiredLevel == "advanced":
			gap = "severe"
		case currentLevel == "beginner" && requiredLevel == "intermediate":
			gap = "moderate"
		case currentLevel == "intermediate" && requiredLevel == "advanced":
			gap = "moderate"
		case currentLevel == "advanced" && requiredLevel == "beginner":
			gap = "none" // Overqualified
		default:
			gap = "minor"
		}
	}
	
	learningPath := []string{}
	estimatedHours := 0
	resources := []string{}
	
	switch gap {
	case "severe":
		learningPath = []string{
			"Learn fundamentals of " + skillName,
			"Practice with small projects",
			"Build a portfolio project",
			"Get real-world experience",
		}
		estimatedHours = 100
		resources = []string{
			"https://www.coursera.org/search?query=" + skillName,
			"https://www.udemy.com/courses/search/?q=" + skillName,
			"https://www.youtube.com/results?search_query=" + skillName + "+tutorial",
		}
	case "moderate":
		learningPath = []string{
			"Deepen understanding of " + skillName,
			"Work on advanced projects",
			"Contribute to open source",
		}
		estimatedHours = 40
		resources = []string{
			"https://www.coursera.org/search?query=advanced+" + skillName,
			"https://github.com/topics/" + strings.ToLower(skillName),
		}
	default:
		learningPath = []string{
			"Stay updated with latest trends",
			"Consider teaching others",
		}
		estimatedHours = 10
		resources = []string{
			"https://news.ycombinator.com",
			"https://dev.to/t/" + strings.ToLower(skillName),
		}
	}
	
	return &SkillGapReport{
		SkillName:      skillName,
		CurrentLevel:   currentLevel,
		RequiredLevel:  requiredLevel,
		Gap:            gap,
		LearningPath:   learningPath,
		EstimatedHours: estimatedHours,
		Resources:      resources,
	}
}