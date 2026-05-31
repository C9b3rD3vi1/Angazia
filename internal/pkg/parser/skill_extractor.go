package parser

import (
	"strings"
)

type SkillExtractor struct {
	skillDatabase map[string][]string // category -> skills
}

func NewSkillExtractor() *SkillExtractor {
	return &SkillExtractor{
		skillDatabase: buildSkillDatabase(),
	}
}

func (s *SkillExtractor) ExtractSkills(text string) ([]string, error) {
	textLower := strings.ToLower(text)
	foundSkills := make(map[string]bool)
	
	// Check each skill category
	for _, skills := range s.skillDatabase {
		for _, skill := range skills {
			if strings.Contains(textLower, strings.ToLower(skill)) {
				foundSkills[skill] = true
			}
		}
	}
	
	// Convert map to slice
	skills := make([]string, 0, len(foundSkills))
	for skill := range foundSkills {
		skills = append(skills, skill)
	}
	
	return skills, nil
}

// ExtractSkillsWithCategories returns skills grouped by category
func (s *SkillExtractor) ExtractSkillsWithCategories(text string) (map[string][]string, error) {
	textLower := strings.ToLower(text)
	foundCategories := make(map[string][]string)
	
	for category, skills := range s.skillDatabase {
		for _, skill := range skills {
			if strings.Contains(textLower, strings.ToLower(skill)) {
				foundCategories[category] = append(foundCategories[category], skill)
			}
		}
	}
	
	return foundCategories, nil
}

// GetTopSkillCategories returns the most relevant skill categories based on matches
func (s *SkillExtractor) GetTopSkillCategories(text string, limit int) ([]string, error) {
	categoryCounts := make(map[string]int)
	textLower := strings.ToLower(text)
	
	for category, skills := range s.skillDatabase {
		count := 0
		for _, skill := range skills {
			if strings.Contains(textLower, strings.ToLower(skill)) {
				count++
			}
		}
		if count > 0 {
			categoryCounts[category] = count
		}
	}
	
	// Sort by count (higher first)
	type categoryCount struct {
		category string
		count    int
	}
	
	sorted := make([]categoryCount, 0, len(categoryCounts))
	for cat, cnt := range categoryCounts {
		sorted = append(sorted, categoryCount{category: cat, count: cnt})
	}
	
	// Bubble sort by count descending
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	limit = min(limit, len(sorted))
	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, sorted[i].category)
	}
	
	return result, nil
}

// GetMissingSkillsForCategory returns skills a candidate is missing for a specific category
func (s *SkillExtractor) GetMissingSkillsForCategory(text string, category string) ([]string, error) {
	textLower := strings.ToLower(text)
	category = strings.ToLower(category)
	
	skills, ok := s.skillDatabase[category]
	if !ok {
		return []string{}, nil
	}
	
	missingSkills := make([]string, 0)
	for _, skill := range skills {
		if !strings.Contains(textLower, strings.ToLower(skill)) {
			missingSkills = append(missingSkills, skill)
		}
	}
	
	return missingSkills, nil
}

// SkillRecommendation represents a recommended skill with its category
type SkillRecommendation struct {
	Skill    string `json:"skill"`
	Category string `json:"category"`
	Reason   string `json:"reason"`
}

// GetSkillRecommendationsWithCategories returns skill recommendations with category info
func (s *SkillExtractor) GetSkillRecommendationsWithCategories(existingSkills []string, limit int) ([]SkillRecommendation, error) {
	recommended := make(map[string]SkillRecommendation)
	existingLower := make(map[string]bool)
	
	for _, skill := range existingSkills {
		existingLower[strings.ToLower(skill)] = true
	}
	
	// Find related skills from same categories
	for category, skills := range s.skillDatabase {
		hasCategorySkill := false
		categorySkills := make([]string, 0)
		
		for _, skill := range skills {
			if existingLower[strings.ToLower(skill)] {
				hasCategorySkill = true
			}
			categorySkills = append(categorySkills, skill)
		}
		
		if hasCategorySkill {
			// Add other skills from this category
			for _, skill := range categorySkills {
				if !existingLower[strings.ToLower(skill)] {
					if _, exists := recommended[skill]; !exists {
						recommended[skill] = SkillRecommendation{
							Skill:    skill,
							Category: category,
							Reason:   "Related to your existing " + category + " skills",
						}
					}
				}
			}
		}
	}
	
	// Also add complementary skills
	complementarySkills := map[string][]ComplementSkill{
		"go": {
			{Skill: "docker", Category: "devops", Reason: "Essential for containerizing Go applications"},
			{Skill: "kubernetes", Category: "devops", Reason: "For orchestrating Go microservices"},
			{Skill: "grpc", Category: "backend", Reason: "High-performance RPC framework for Go"},
			{Skill: "postgresql", Category: "database", Reason: "Common database with Go applications"},
		},
		"python": {
			{Skill: "django", Category: "backend", Reason: "Popular web framework for Python"},
			{Skill: "flask", Category: "backend", Reason: "Lightweight Python web framework"},
			{Skill: "fastapi", Category: "backend", Reason: "Modern Python web framework"},
			{Skill: "pandas", Category: "data", Reason: "Essential for data analysis in Python"},
		},
		"react": {
			{Skill: "typescript", Category: "frontend", Reason: "Adds type safety to React apps"},
			{Skill: "redux", Category: "frontend", Reason: "State management for React"},
			{Skill: "next.js", Category: "frontend", Reason: "React framework for production"},
			{Skill: "tailwind", Category: "frontend", Reason: "Utility-first CSS for React"},
		},
		"javascript": {
			{Skill: "typescript", Category: "frontend", Reason: "Superset of JavaScript with types"},
			{Skill: "react", Category: "frontend", Reason: "UI library for web apps"},
			{Skill: "node.js", Category: "backend", Reason: "JavaScript runtime for backend"},
			{Skill: "webpack", Category: "frontend", Reason: "Module bundler for JS apps"},
		},
		"aws": {
			{Skill: "terraform", Category: "devops", Reason: "Infrastructure as Code for AWS"},
			{Skill: "kubernetes", Category: "devops", Reason: "Container orchestration on AWS"},
			{Skill: "serverless", Category: "cloud", Reason: "AWS Lambda and serverless architecture"},
			{Skill: "cloudformation", Category: "cloud", Reason: "AWS native IaC tool"},
		},
		"docker": {
			{Skill: "kubernetes", Category: "devops", Reason: "Container orchestration"},
			{Skill: "jenkins", Category: "devops", Reason: "CI/CD pipeline automation"},
			{Skill: "terraform", Category: "devops", Reason: "Infrastructure as Code"},
			{Skill: "prometheus", Category: "devops", Reason: "Monitoring for containers"},
		},
	}
	
	for _, skill := range existingSkills {
		skillLower := strings.ToLower(skill)
		if comps, ok := complementarySkills[skillLower]; ok {
			for _, comp := range comps {
				if !existingLower[strings.ToLower(comp.Skill)] {
					if _, exists := recommended[comp.Skill]; !exists {
						recommended[comp.Skill] = SkillRecommendation{
							Skill:    comp.Skill,
							Category: comp.Category,
							Reason:   comp.Reason,
						}
					}
				}
			}
		}
	}
	
	// Convert to slice and sort by relevance (priority to skills from categories where user has most skills)
	result := make([]SkillRecommendation, 0, len(recommended))
	for _, rec := range recommended {
		result = append(result, rec)
	}
	
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	
	return result, nil
}

// ComplementSkill represents a complementary skill suggestion
type ComplementSkill struct {
	Skill    string
	Category string
	Reason   string
}

// GetSkillRecommendations suggests skills based on existing skills (legacy method for backward compatibility)
func (s *SkillExtractor) GetSkillRecommendations(existingSkills []string, limit int) ([]string, error) {
	recommendations, err := s.GetSkillRecommendationsWithCategories(existingSkills, limit)
	if err != nil {
		return nil, err
	}
	
	result := make([]string, len(recommendations))
	for i, rec := range recommendations {
		result[i] = rec.Skill
	}
	
	return result, nil
}

// GetCategorySummary returns a summary of skills by category
func (s *SkillExtractor) GetCategorySummary(text string) (map[string]int, error) {
	textLower := strings.ToLower(text)
	summary := make(map[string]int)
	
	for category, skills := range s.skillDatabase {
		count := 0
		for _, skill := range skills {
			if strings.Contains(textLower, strings.ToLower(skill)) {
				count++
			}
		}
		if count > 0 {
			summary[category] = count
		}
	}
	
	return summary, nil
}

func (s *SkillExtractor) GetSkillCategories() []string {
	categories := make([]string, 0, len(s.skillDatabase))
	for category := range s.skillDatabase {
		categories = append(categories, category)
	}
	return categories
}

func (s *SkillExtractor) GetSkillsByCategory(category string) []string {
	if skills, ok := s.skillDatabase[strings.ToLower(category)]; ok {
		return skills
	}
	return []string{}
}

// GetCategoryBySkill returns the category a skill belongs to
func (s *SkillExtractor) GetCategoryBySkill(skill string) string {
	skillLower := strings.ToLower(skill)
	for category, skills := range s.skillDatabase {
		for _, catSkill := range skills {
			if strings.ToLower(catSkill) == skillLower {
				return category
			}
		}
	}
	return ""
}

// GetAllSkills returns all skills from all categories
func (s *SkillExtractor) GetAllSkills() []string {
	allSkills := make([]string, 0)
	for _, skills := range s.skillDatabase {
		allSkills = append(allSkills, skills...)
	}
	return allSkills
}

// GetSkillCount returns total number of skills in database
func (s *SkillExtractor) GetSkillCount() int {
	count := 0
	for _, skills := range s.skillDatabase {
		count += len(skills)
	}
	return count
}

func buildSkillDatabase() map[string][]string {
	return map[string][]string{
		"backend": {
			"go", "golang", "python", "java", "node.js", "nodejs", "php", "ruby", "c#", "c++",
			"rust", "scala", "kotlin", "spring boot", "django", "flask", "fastapi", "gin",
			"echo", "fiber", "laravel", "rails", "express.js", "nestjs",
		},
		"frontend": {
			"react", "react.js", "reactjs", "vue", "vue.js", "angular", "angularjs", "svelte",
			"next.js", "nextjs", "gatsby", "html5", "css3", "javascript", "typescript", "tailwind",
			"bootstrap", "material ui", "jquery", "redux", "vuex", "webpack", "vite",
		},
		"mobile": {
			"react native", "flutter", "swift", "kotlin", "android", "ios", "ionic", "xamarin",
			"java android", "objective-c", "dart", "mobile development",
		},
		"database": {
			"postgresql", "postgres", "mysql", "mongodb", "redis", "elasticsearch", "cassandra",
			"dynamodb", "firebase", "sqlite", "mariadb", "oracle", "sql server", "influxdb",
		},
		"devops": {
			"docker", "kubernetes", "k8s", "jenkins", "gitlab ci", "github actions", "terraform",
			"ansible", "puppet", "chef", "aws", "azure", "gcp", "linux", "nginx", "apache",
			"prometheus", "grafana", "elk", "splunk", "ci/cd", "devops",
		},
		"cloud": {
			"aws", "amazon web services", "ec2", "s3", "lambda", "cloudfront", "rds", "dynamodb",
			"azure", "microsoft azure", "gcp", "google cloud", "heroku", "digitalocean", "linode",
			"cloud computing", "serverless", "cloudformation", "cdk",
		},
		"data": {
			"python", "pandas", "numpy", "tensorflow", "pytorch", "scikit-learn", "machine learning",
			"data science", "data analysis", "sql", "etl", "data engineering", "apache spark",
			"hadoop", "airflow", "dbt", "looker", "tableau", "power bi",
		},
		"security": {
			"cybersecurity", "penetration testing", "vulnerability assessment", "owasp", "burp suite",
			"metasploit", "nmap", "wireshark", "firewall", "ids", "ips", "soc", "siem", "compliance",
			"iso 27001", "gdpr", "security+", "cissp", "ceh", "oscp",
		},
		"soft_skills": {
			"leadership", "communication", "teamwork", "problem solving", "critical thinking",
			"project management", "agile", "scrum", "kanban", "jira", "confluence", "mentoring",
			"presentation", "negotiation", "time management",
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}