package parser

import (
	"regexp"
	"strconv"
	"strings"
)

type ExperienceExtractor struct {
	yearRegex *regexp.Regexp
}

func NewExperienceExtractor() *ExperienceExtractor {
	return &ExperienceExtractor{
		yearRegex: regexp.MustCompile(`20\d{2}`),
	}
}

func (e *ExperienceExtractor) ExtractExperience(text string) (*ExperienceInfo, error) {
	info := &ExperienceInfo{
		CareerProgression: []string{},
		Roles:             []string{},
		Industries:        []string{},
		TopCompanies:      []string{},
	}
	
	// Extract years of experience
	info.TotalYears = e.extractTotalYears(text)
	
	// Determine seniority level
	info.SeniorityLevel = e.determineSeniorityLevel(text, info.TotalYears)
	
	// Extract roles
	info.Roles = e.extractRoles(text)
	
	// Extract industries
	info.Industries = e.extractIndustries(text)
	
	return info, nil
}

func (e *ExperienceExtractor) extractTotalYears(text string) int {
	totalYears := 0
	
	// Pattern 1: "X years of experience" or "X+ years"
	pattern1 := regexp.MustCompile(`(\d+)\s*(?:\+)?\s*years?\s*(?:of)?\s*experience`)
	if matches := pattern1.FindStringSubmatch(strings.ToLower(text)); len(matches) > 1 {
		if years, err := strconv.Atoi(matches[1]); err == nil && years > totalYears {
			totalYears = years
		}
	}
	
	// Pattern 2: "experience of X years"
	pattern2 := regexp.MustCompile(`experience\s*(?:of)?\s*(\d+)\s*(?:\+)?\s*years?`)
	if matches := pattern2.FindStringSubmatch(strings.ToLower(text)); len(matches) > 1 {
		if years, err := strconv.Atoi(matches[1]); err == nil && years > totalYears {
			totalYears = years
		}
	}
	
	// Pattern 3: "X+ years exp"
	pattern3 := regexp.MustCompile(`(\d+)\s*(?:\+)?\s*years?\s+exp`)
	if matches := pattern3.FindStringSubmatch(strings.ToLower(text)); len(matches) > 1 {
		if years, err := strconv.Atoi(matches[1]); err == nil && years > totalYears {
			totalYears = years
		}
	}
	
	// If no explicit years, try to calculate from dates
	if totalYears == 0 {
		totalYears = e.calculateYearsFromDates(text)
	}
	
	return totalYears
}

func (e *ExperienceExtractor) calculateYearsFromDates(text string) int {
	datePattern := regexp.MustCompile(`(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec|20\d{2})\s*(?:-|–|to)\s*(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec|Present|Current|20\d{2})`)
	matches := datePattern.FindAllString(text, -1)
	
	totalMonths := 0
	for _, match := range matches {
		parts := strings.Split(match, "-")
		if len(parts) == 2 {
			startYear := e.extractYearFromString(parts[0])
			endYear := e.extractYearFromString(parts[1])
			if startYear > 0 && endYear > 0 {
				totalMonths += (endYear - startYear) * 12
			}
		}
	}
	
	// Also count current positions (Present)
	if strings.Contains(strings.ToLower(text), "present") {
		// Add 6 months as estimate for current role
		totalMonths += 6
	}
	
	return totalMonths / 12
}

func (e *ExperienceExtractor) extractYearFromString(s string) int {
	match := e.yearRegex.FindString(s)
	if match != "" {
		if year, err := strconv.Atoi(match); err == nil {
			return year
		}
	}
	return 0
}

func (e *ExperienceExtractor) determineSeniorityLevel(text string, totalYears int) string {
	textLower := strings.ToLower(text)
	
	// Check for explicit titles
	if strings.Contains(textLower, "senior") || strings.Contains(textLower, "lead") ||
	   strings.Contains(textLower, "principal") || strings.Contains(textLower, "staff") ||
	   strings.Contains(textLower, "architect") {
		if totalYears >= 5 {
			return "senior"
		}
	}
	
	if strings.Contains(textLower, "mid") || strings.Contains(textLower, "intermediate") {
		return "mid"
	}
	
	if strings.Contains(textLower, "junior") || strings.Contains(textLower, "entry") ||
	   strings.Contains(textLower, "associate") {
		return "entry"
	}
	
	// Fallback to years-based
	if totalYears >= 7 {
		return "senior"
	} else if totalYears >= 3 {
		return "mid"
	} else if totalYears >= 1 {
		return "junior"
	}
	
	return "entry"
}

func (e *ExperienceExtractor) extractRoles(text string) []string {
	roles := make(map[string]bool)
	
	rolePattern := regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\s+(?:Engineer|Developer|Architect|Lead|Manager|Director|Consultant|Analyst|Administrator|Specialist|Coordinator))`)
	
	matches := rolePattern.FindAllString(text, -1)
	for _, match := range matches {
		if len(match) > 3 && len(match) < 60 {
			roles[strings.TrimSpace(match)] = true
		}
	}
	
	// Also look for role keywords in lines
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		for _, keyword := range []string{"engineer", "developer", "architect", "lead", "manager"} {
			if strings.Contains(strings.ToLower(line), keyword) {
				cleaned := strings.TrimSpace(line)
				if len(cleaned) > 5 && len(cleaned) < 80 {
					roles[cleaned] = true
				}
				break
			}
		}
	}
	
	result := make([]string, 0, len(roles))
	for role := range roles {
		result = append(result, role)
	}
	
	return result
}

func (e *ExperienceExtractor) extractIndustries(text string) []string {
	industries := make(map[string]bool)
	
	industryKeywords := map[string][]string{
		"fintech":      {"bank", "finance", "payment", "mobile money", "mpesa", "financial", "insurance"},
		"ecommerce":    {"ecommerce", "retail", "marketplace", "shop", "store", "commerce"},
		"healthtech":   {"health", "medical", "clinic", "hospital", "wellness", "healthcare"},
		"edtech":       {"education", "learning", "training", "school", "academy", "university"},
		"logistics":    {"logistics", "delivery", "transport", "shipping", "courier", "supply chain"},
		"saas":         {"saas", "software as a service", "cloud software", "subscription"},
		"telecom":      {"telecom", "communication", "mobile network", "isp", "telco"},
		"consulting":   {"consulting", "consultancy", "advisory", "professional services"},
		"government":   {"government", "public sector", "ministry", "state", "parastatal"},
		"startup":      {"startup", "venture", "incubator", "accelerator"},
	}
	
	textLower := strings.ToLower(text)
	
	for industry, keywords := range industryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(textLower, keyword) {
				industries[industry] = true
				break
			}
		}
	}
	
	result := make([]string, 0, len(industries))
	for industry := range industries {
		result = append(result, industry)
	}
	
	return result
}