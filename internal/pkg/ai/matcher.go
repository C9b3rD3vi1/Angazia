package ai

import (
	"fmt"
	"math"
	"strings"
)

// Matcher provides additional matching utilities beyond AI providers
type Matcher struct {
	skillSynonyms map[string][]string
	industryKeywords map[string][]string
}

func NewMatcher() *Matcher {
	return &Matcher{
		skillSynonyms:   buildSkillSynonyms(),
		industryKeywords: buildIndustryKeywords(),
	}
}

// CalculateSkillMatch calculates skill match percentage between job and candidate
func (m *Matcher) CalculateSkillMatch(jobSkills, candidateSkills []string) (matchingSkills, missingSkills []string, percentage int) {
	candidateSkillSet := make(map[string]bool)
	for _, s := range candidateSkills {
		normalized := normalizeSkill(s)
		candidateSkillSet[normalized] = true
		
		// Add synonyms
		if synonyms, ok := m.skillSynonyms[normalized]; ok {
			for _, syn := range synonyms {
				candidateSkillSet[syn] = true
			}
		}
	}
	
	matchingMap := make(map[string]bool)
	
	for _, jobSkill := range jobSkills {
		normalized := normalizeSkill(jobSkill)
		if candidateSkillSet[normalized] {
			matchingMap[normalized] = true
			matchingSkills = append(matchingSkills, jobSkill)
		} else {
			missingSkills = append(missingSkills, jobSkill)
		}
	}
	
	if len(jobSkills) > 0 {
		percentage = (len(matchingSkills) * 100) / len(jobSkills)
	}
	
	return matchingSkills, missingSkills, percentage
}

// CalculateExperienceMatch calculates experience match
func (m *Matcher) CalculateExperienceMatch(jobMinExp, jobMaxExp, candidateYears int) int {
	if candidateYears <= 0 {
		return 0
	}
	
	if jobMinExp > 0 && candidateYears >= jobMinExp {
		if jobMaxExp > 0 && candidateYears <= jobMaxExp {
			return 100
		}
		return 80
	}
	
	if jobMinExp > 0 {
		ratio := float64(candidateYears) / float64(jobMinExp)
		return int(math.Min(ratio*100, 100))
	}
	
	return 50
}

// CalculateLocationMatch calculates location match
func (m *Matcher) CalculateLocationMatch(jobLocation string, isRemote bool, candidateLocation string, isRemoteOnly bool) int {
	if isRemote && isRemoteOnly {
		return 100
	}
	
	if isRemote && !isRemoteOnly {
		return 80
	}
	
	if !isRemote && strings.EqualFold(jobLocation, candidateLocation) {
		return 100
	}
	
	if strings.Contains(strings.ToLower(jobLocation), strings.ToLower(candidateLocation)) ||
		strings.Contains(strings.ToLower(candidateLocation), strings.ToLower(jobLocation)) {
		return 60
	}
	
	return 20
}

// CalculateCultureMatch calculates culture match based on GitHub activity and profile completeness
func (m *Matcher) CalculateCultureMatch(githubActivity *GithubActivity, profileCompleteness int) int {
	score := 50 // Base score
	
	if githubActivity != nil {
		if githubActivity.ActivityScore >= 80 {
			score += 20
		} else if githubActivity.ActivityScore >= 60 {
			score += 15
		} else if githubActivity.ActivityScore >= 40 {
			score += 10
		} else if githubActivity.ActivityScore > 0 {
			score += 5
		}
		
		if githubActivity.ContributionStreak > 30 {
			score += 10
		} else if githubActivity.ContributionStreak > 14 {
			score += 5
		}
	}
	
	if profileCompleteness >= 80 {
		score += 15
	} else if profileCompleteness >= 60 {
		score += 10
	} else if profileCompleteness >= 40 {
		score += 5
	}
	
	if score > 100 {
		score = 100
	}
	
	return score
}

// CalculateOverallScore calculates weighted overall score
func (m *Matcher) CalculateOverallScore(skillsScore, experienceScore, locationScore, cultureScore int, weights *MatchWeights) int {
	if weights == nil {
		weights = DefaultMatchWeights()
	}
	
	total := (skillsScore * weights.Skills) +
		(experienceScore * weights.Experience) +
		(locationScore * weights.Location) +
		(cultureScore * weights.Culture)
	
	return total / 100
}

// GetRecommendation returns hiring recommendation based on score
func (m *Matcher) GetRecommendation(score int) string {
	switch {
	case score >= 85:
		return "hire"
	case score >= 70:
		return "interview"
	case score >= 50:
		return "consider"
	default:
		return "reject"
	}
}

// GenerateMatchSummary creates a human-readable match summary
func (m *Matcher) GenerateMatchSummary(jobTitle, candidateName string, score int, matchingSkills, missingSkills []string) string {
	var summary strings.Builder
	
	summary.WriteString(fmt.Sprintf("%s matches the %s position with a %d%% compatibility score. ", candidateName, jobTitle, score))
	
	if len(matchingSkills) > 0 {
		summary.WriteString(fmt.Sprintf("They have strong alignment in: %s. ", strings.Join(matchingSkills[:min(3, len(matchingSkills))], ", ")))
	}
	
	if len(missingSkills) > 0 {
		summary.WriteString(fmt.Sprintf("Skill gaps identified: %s. ", strings.Join(missingSkills[:min(3, len(missingSkills))], ", ")))
	}
	
	if score >= 80 {
		summary.WriteString("This candidate is an excellent fit and should be prioritized for an interview.")
	} else if score >= 60 {
		summary.WriteString("This candidate is a good fit with some areas for improvement.")
	} else {
		summary.WriteString("This candidate may require additional training or may be better suited for other positions.")
	}
	
	return summary.String()
}

// MatchWeights defines weights for different match components
type MatchWeights struct {
	Skills     int
	Experience int
	Location   int
	Culture    int
}

func DefaultMatchWeights() *MatchWeights {
	return &MatchWeights{
		Skills:     50,
		Experience: 25,
		Location:   15,
		Culture:    10,
	}
}

// Helper functions
func normalizeSkill(skill string) string {
	normalized := strings.ToLower(strings.TrimSpace(skill))
	
	// Remove common suffixes
	suffixes := []string{"js", "jsx", "ts", "tsx", "framework", "library", "language"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
		}
	}
	
	// Handle common variations
	variations := map[string]string{
		"react.js":     "react",
		"reactjs":      "react",
		"node.js":      "nodejs",
		"nodejs":       "nodejs",
		"postgresql":   "postgres",
		"typescript":   "typescript",
		"javascript":   "javascript",
		"golang":       "go",
		"golanglang":   "go",
	}
	
	if val, ok := variations[normalized]; ok {
		return val
	}
	
	return normalized
}

func buildSkillSynonyms() map[string][]string {
	return map[string][]string{
		"react":       {"reactjs", "react.js", "react-js", "react framework"},
		"vue":         {"vuejs", "vue.js", "vue-js"},
		"angular":     {"angularjs", "angular.js", "angular-js"},
		"nodejs":      {"node", "node.js", "node-js"},
		"python":      {"py", "python3", "python-3"},
		"golang":      {"go", "golanglang"},
		"postgres":    {"postgresql", "pg"},
		"mongodb":     {"mongo", "mongodb-community"},
		"redis":       {"redis-cache", "redis-db"},
		"docker":      {"docker-container", "docker-engine"},
		"kubernetes":  {"k8s", "kube"},
		"aws":         {"amazon-web-services", "amazon aws"},
		"azure":       {"microsoft-azure", "ms-azure"},
		"gcp":         {"google-cloud", "google-cloud-platform"},
	}
}

func buildIndustryKeywords() map[string][]string {
	return map[string][]string{
		"fintech":    {"financial", "banking", "payment", "mobile money", "mpesa"},
		"ecommerce":  {"retail", "marketplace", "shopping", "store"},
		"healthtech": {"healthcare", "medical", "wellness", "hospital"},
		"edtech":     {"education", "learning", "training", "academy"},
		"logistics":  {"delivery", "transport", "shipping", "supply chain"},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}