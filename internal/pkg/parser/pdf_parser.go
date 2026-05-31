package parser

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"mime/multipart"
	"unicode"

	"github.com/ledongthuc/pdf"
)

type PDFParser struct {
	skillExtractor      *SkillExtractor
	experienceExtractor *ExperienceExtractor
	emailRegex          *regexp.Regexp
	phoneRegex          *regexp.Regexp
	linkedinRegex       *regexp.Regexp
	githubRegex         *regexp.Regexp
	twitterRegex        *regexp.Regexp
	urlRegex            *regexp.Regexp
}

func NewPDFParser() *PDFParser {
	return &PDFParser{
		skillExtractor:      NewSkillExtractor(),
		experienceExtractor: NewExperienceExtractor(),
		emailRegex:          regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		phoneRegex:          regexp.MustCompile(`(\+?254|0|01)?[0-9]{9,12}`),
		linkedinRegex:       regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?linkedin\.com\/in\/[a-zA-Z0-9_-]+`),
		githubRegex:         regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?github\.com\/[a-zA-Z0-9_-]+`),
		twitterRegex:        regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?twitter\.com\/[a-zA-Z0-9_-]+`),
		urlRegex:            regexp.MustCompile(`https?:\/\/[^\s]+`),
	}
}

func (p *PDFParser) Parse(ctx context.Context, file multipart.File, filename string) (*ParsedResume, error) {
	// Extract text from PDF
	content, err := p.extractText(file)
	if err != nil {
		return nil, fmt.Errorf("failed to extract PDF text: %w", err)
	}
	
	// Clean and normalize text
	content = p.normalizeText(content)
	
	// Parse sections
	resume := &ParsedResume{}
	
	// Extract contact info
	contactInfo, err := p.ExtractContactInfo(ctx, content)
	if err == nil && contactInfo != nil {
		resume.Email = contactInfo.Email
		resume.Phone = contactInfo.Phone
		resume.LinkedInURL = contactInfo.LinkedIn
		resume.PortfolioURL = contactInfo.Website
	}
	
	// Extract name (usually first line or after common headers)
	resume.FullName = p.extractName(content)
	
	// Extract location
	resume.Location = p.extractLocation(content)
	
	// Extract skills
	resume.Skills, _ = p.ExtractSkills(ctx, content)
	
	// Extract experience
	experienceInfo, experienceList, _ := p.ExtractExperience(ctx, content)
	if experienceInfo != nil {
		resume.TotalExperience = experienceInfo.TotalYears
	}
	resume.Experience = experienceList
	
	// Extract education
	resume.Education, _ = p.ExtractEducation(ctx, content)
	
	// Extract summary
	resume.Summary = p.extractSummary(content)
	
	// Extract certifications
	resume.Certifications = p.extractCertifications(content)
	
	// Extract languages
	resume.Languages = p.extractLanguages(content)
	
	return resume, nil
}

func (p *PDFParser) extractText(file multipart.File) (string, error) {
	// Save file to temp buffer
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		return "", err
	}
	
	// Parse PDF
	reader := bytes.NewReader(buf.Bytes())
	pdfReader, err := pdf.NewReader(reader, reader.Size())
	if err != nil {
		return "", err
	}
	
	var text strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text.WriteString(content)
		text.WriteString("\n")
	}
	
	return text.String(), nil
}

func (p *PDFParser) normalizeText(text string) string {
	// Replace multiple newlines with single
	text = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(text, "\n")
	
	// Replace multiple spaces with single
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	
	// Remove non-printable characters
	text = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return -1
	}, text)
	
	return strings.TrimSpace(text)
}

func (p *PDFParser) extractName(text string) string {
	lines := strings.Split(text, "\n")
	
	// Check first few lines for name
	for i := 0; i < 5 && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		// Skip lines that look like headers
		lowerLine := strings.ToLower(line)
		skipPatterns := []string{"resume", "curriculum vitae", "cv", "contact", "email", "phone"}
		shouldSkip := false
		for _, pattern := range skipPatterns {
			if strings.Contains(lowerLine, pattern) {
				shouldSkip = true
				break
			}
		}
		
		if !shouldSkip && len(line) < 50 && !p.containsDigit(line) {
			// Check if it looks like a name (has spaces, proper case)
			words := strings.Fields(line)
			if len(words) >= 2 && len(words) <= 4 {
				return line
			}
		}
	}
	
	return ""
}

func (p *PDFParser) containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func (p *PDFParser) extractLocation(text string) string {
	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		
		// Look for location indicators
		if strings.Contains(lowerLine, "location") || 
		   strings.Contains(lowerLine, "based in") ||
		   strings.Contains(lowerLine, "living in") {
			
			// Extract after colon
			if idx := strings.Index(line, ":"); idx != -1 && idx+1 < len(line) {
				location := strings.TrimSpace(line[idx+1:])
				if len(location) > 3 && len(location) < 50 {
					return location
				}
			}
		}
	}
	
	// Check for Kenyan cities
	kenyanCities := []string{"nairobi", "mombasa", "kisumu", "nakuru", "eldoret", "thika", "malindi"}
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		for _, city := range kenyanCities {
			if strings.Contains(lowerLine, city) {
				// Extract the city with proper case
				for _, word := range strings.Fields(line) {
					if strings.ToLower(word) == city {
						return word
					}
				}
				return city
			}
		}
	}
	
	return ""
}

func (p *PDFParser) extractSummary(text string) string {
	lines := strings.Split(text, "\n")
	
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		
		// Look for summary headers
		summaryHeaders := []string{"summary", "profile", "about me", "professional summary", "career objective"}
		for _, header := range summaryHeaders {
			if strings.Contains(lowerLine, header) {
				// Get next 2-3 lines as summary
				summaryLines := []string{}
				for j := i + 1; j < len(lines) && j <= i+5; j++ {
					nextLine := strings.TrimSpace(lines[j])
					if nextLine == "" {
						continue
					}
					// Stop if we hit another section header
					if p.isSectionHeader(nextLine) {
						break
					}
					summaryLines = append(summaryLines, nextLine)
					if len(strings.Join(summaryLines, " ")) > 300 {
						break
					}
				}
				
				if len(summaryLines) > 0 {
					return strings.Join(summaryLines, " ")
				}
			}
		}
	}
	
	return ""
}

func (p *PDFParser) isSectionHeader(line string) bool {
	lowerLine := strings.ToLower(line)
	headers := []string{"experience", "education", "skills", "certifications", "projects", "languages"}
	for _, header := range headers {
		if strings.Contains(lowerLine, header) && len(line) < 30 {
			return true
		}
	}
	return false
}

func (p *PDFParser) extractCertifications(text string) []string {
	certifications := []string{}
	certKeywords := []string{"certified", "certification", "certificate", "scrum", "aws certified", "azure certified"}
	
	lines := strings.Split(text, "\n")
	inCertSection := false
	
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		
		if strings.Contains(lowerLine, "certification") {
			inCertSection = true
			continue
		}
		
		if inCertSection {
			if p.isSectionHeader(line) && !strings.Contains(lowerLine, "cert") {
				inCertSection = false
				continue
			}
			
			for _, keyword := range certKeywords {
				if strings.Contains(lowerLine, keyword) {
					cert := strings.TrimSpace(line)
					if len(cert) > 3 && len(cert) < 100 {
						certifications = append(certifications, cert)
						break
					}
				}
			}
			
			// Also check next line for certification details
			if i+1 < len(lines) && len(certifications) > 0 {
				nextLine := strings.TrimSpace(lines[i+1])
				if len(nextLine) > 10 && len(nextLine) < 100 && !p.isSectionHeader(nextLine) {
					certifications[len(certifications)-1] = certifications[len(certifications)-1] + " " + nextLine
				}
			}
		}
	}
	
	return certifications
}

func (p *PDFParser) extractLanguages(text string) []string {
	languages := []string{}
	
	languageList := []string{
		"english", "swahili", "kiswahili", "french", "german", "spanish", 
		"mandarin", "arabic", "kikuyu", "luo", "kalenjin", "kamba", "meru",
	}
	
	lines := strings.Split(text, "\n")
	inLangSection := false
	
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		
		if strings.Contains(lowerLine, "language") {
			inLangSection = true
			continue
		}
		
		if inLangSection {
			if p.isSectionHeader(line) {
				inLangSection = false
				continue
			}
			
			for _, lang := range languageList {
				if strings.Contains(lowerLine, lang) {
					languages = append(languages, lang)
				}
			}
		}
	}
	
	return languages
}

func (p *PDFParser) ExtractSkills(ctx context.Context, text string) ([]string, error) {
	return p.skillExtractor.ExtractSkills(text)
}

func (p *PDFParser) ExtractExperience(ctx context.Context, text string) (*ExperienceInfo, []Experience, error) {
	info, err := p.experienceExtractor.ExtractExperience(text)
	
	// Also extract structured experience
	experiences := p.extractStructuredExperience(text)
	
	return info, experiences, err
}

func (p *PDFParser) extractStructuredExperience(text string) []Experience {
	experiences := []Experience{}
	
	lines := strings.Split(text, "\n")
	inExpSection := false
	currentExp := Experience{}
	
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)
		
		// Detect experience section
		if strings.Contains(lowerLine, "experience") || 
		   strings.Contains(lowerLine, "work history") ||
		   strings.Contains(lowerLine, "employment") {
			inExpSection = true
			continue
		}
		
		if !inExpSection {
			continue
		}
		
		// Exit experience section
		if p.isSectionHeader(line) && !strings.Contains(lowerLine, "experience") {
			break
		}
		
		// Detect job title (usually at start of line, proper case, not too long)
		if len(line) > 3 && len(line) < 60 && !p.isDate(line) && !p.isCompany(line) {
			if currentExp.Title != "" {
				// Save previous experience
				if currentExp.Title != "" || currentExp.Company != "" {
					experiences = append(experiences, currentExp)
				}
				currentExp = Experience{}
			}
			currentExp.Title = line
			continue
		}
		
		// Detect company (often after job title or contains Inc, Ltd, etc.)
		if p.isCompany(line) && currentExp.Company == "" {
			currentExp.Company = line
			continue
		}
		
		// Detect date range
		if p.isDate(line) {
			dateParts := strings.Split(line, "-")
			if len(dateParts) == 2 {
				currentExp.StartDate = strings.TrimSpace(dateParts[0])
				currentExp.EndDate = strings.TrimSpace(dateParts[1])
				if strings.Contains(strings.ToLower(currentExp.EndDate), "present") {
					currentExp.Current = true
				}
			}
			continue
		}
		
		// Detect description bullets
		if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") || 
		   strings.HasPrefix(line, "*") || (len(line) > 10 && unicode.IsDigit(rune(line[0]))) {
			description := strings.TrimLeft(line, "•-* \t")
			if len(description) > 5 {
				currentExp.Description = append(currentExp.Description, description)
			}
		}
	}
	
	// Add last experience
	if currentExp.Title != "" || currentExp.Company != "" {
		experiences = append(experiences, currentExp)
	}
	
	return experiences
}

func (p *PDFParser) isDate(line string) bool {
	datePatterns := []string{
		`\d{4}\s*-\s*\d{4}`,
		`(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{4}\s*-\s*(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{4}`,
		`\d{4}\s*-\s*(Present|Current)`,
		`(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{4}\s*-\s*(Present|Current)`,
	}
	
	for _, pattern := range datePatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func (p *PDFParser) isCompany(line string) bool {
	companyIndicators := []string{"inc", "ltd", "limited", "corporation", "corp", "company", "co", "technologies", "solutions"}
	lowerLine := strings.ToLower(line)
	
	for _, indicator := range companyIndicators {
		if strings.Contains(lowerLine, indicator) {
			return true
		}
	}
	
	// Check for common company name patterns (capitalized words, 2-4 words)
	words := strings.Fields(line)
	if len(words) >= 2 && len(words) <= 5 {
		capitalCount := 0
		for _, word := range words {
			if len(word) > 0 && unicode.IsUpper(rune(word[0])) {
				capitalCount++
			}
		}
		if capitalCount >= 2 {
			return true
		}
	}
	
	return false
}

func (p *PDFParser) ExtractEducation(ctx context.Context, text string) ([]Education, error) {
	educations := []Education{}
	
	lines := strings.Split(text, "\n")
	inEduSection := false
	
	degreePatterns := []string{
		"bachelor", "master", "phd", "doctorate", "diploma", "certificate",
		"bsc", "msc", "ba", "bs", "ma", "mba", "btech", "mtech",
		"associate", "high school", "secondary", "primary",
	}
	
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		lowerLine := strings.ToLower(line)
		
		if strings.Contains(lowerLine, "education") || strings.Contains(lowerLine, "academic") {
			inEduSection = true
			continue
		}
		
		if !inEduSection {
			continue
		}
		
		if p.isSectionHeader(line) && !strings.Contains(lowerLine, "education") {
			break
		}
		
		edu := Education{}
		
		// Check for degree
		for _, pattern := range degreePatterns {
			if strings.Contains(lowerLine, pattern) {
				edu.Degree = line
				break
			}
		}
		
		// Look for institution (next line often contains school name)
		if edu.Degree != "" && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if len(nextLine) > 3 && !p.isDate(nextLine) {
				edu.Institution = nextLine
				i++
			}
		}
		
		// Extract dates
		if i+1 < len(lines) && p.isDate(lines[i+1]) {
			dateLine := lines[i+1]
			dateParts := strings.Split(dateLine, "-")
			if len(dateParts) == 2 {
				// Extract years from dates
				startYear := p.extractYearFromString(dateParts[0])
				endYear := p.extractYearFromString(dateParts[1])
				edu.StartYear = startYear
				edu.EndYear = endYear
			}
			i++
		}
		
		if edu.Degree != "" || edu.Institution != "" {
			educations = append(educations, edu)
		}
	}
	
	return educations, nil
}

func (p *PDFParser) extractYearFromString(s string) int {
	yearRegex := regexp.MustCompile(`\d{4}`)
	match := yearRegex.FindString(s)
	if match != "" {
		var year int
		fmt.Sscanf(match, "%d", &year)
		return year
	}
	return 0
}

func (p *PDFParser) ExtractContactInfo(ctx context.Context, text string) (*ContactInfo, error) {
	info := &ContactInfo{}
	
	// Extract email using regex
	if match := p.emailRegex.FindString(text); match != "" {
		info.Email = strings.ToLower(match)
	}
	
	// Extract phone using regex
	if matches := p.phoneRegex.FindAllString(text, -1); len(matches) > 0 {
		for _, match := range matches {
			// Clean the phone number
			cleaned := strings.TrimSpace(match)
			cleaned = strings.Trim(cleaned, "()-,.")
			
			// Validate Kenyan format or international
			if len(cleaned) >= 10 && len(cleaned) <= 13 {
				if strings.HasPrefix(cleaned, "07") || 
				   strings.HasPrefix(cleaned, "01") ||
				   strings.HasPrefix(cleaned, "+254") || 
				   strings.HasPrefix(cleaned, "254") {
					info.Phone = cleaned
					break
				}
			}
		}
	}
	
	// Extract LinkedIn URL
	if match := p.linkedinRegex.FindString(text); match != "" {
		info.LinkedIn = match
	}
	
	// Extract GitHub URL
	if match := p.githubRegex.FindString(text); match != "" {
		info.GitHub = match
	}
	
	// Extract Twitter URL
	if match := p.twitterRegex.FindString(text); match != "" {
		info.Twitter = match
	}
	
	// Extract other URLs (potential portfolio)
	if urls := p.urlRegex.FindAllString(text, -1); len(urls) > 0 {
		for _, url := range urls {
			if !strings.Contains(strings.ToLower(url), "linkedin") &&
			   !strings.Contains(strings.ToLower(url), "github") &&
			   !strings.Contains(strings.ToLower(url), "twitter") {
				info.Website = url
				break
			}
		}
	}
	
	return info, nil
}

// Helper method to parse resume from text (useful for testing)
func (p *PDFParser) ParseFromText(ctx context.Context, text string) (*ParsedResume, error) {
	text = p.normalizeText(text)
	
	resume := &ParsedResume{}
	
	// Extract contact info
	contactInfo, err := p.ExtractContactInfo(ctx, text)
	if err == nil && contactInfo != nil {
		resume.Email = contactInfo.Email
		resume.Phone = contactInfo.Phone
		resume.LinkedInURL = contactInfo.LinkedIn
		resume.PortfolioURL = contactInfo.Website
	}
	
	// Extract name
	resume.FullName = p.extractName(text)
	
	// Extract location
	resume.Location = p.extractLocation(text)
	
	// Extract skills
	resume.Skills, _ = p.ExtractSkills(ctx, text)
	
	// Extract experience
	experienceInfo, experienceList, _ := p.ExtractExperience(ctx, text)
	if experienceInfo != nil {
		resume.TotalExperience = experienceInfo.TotalYears
	}
	resume.Experience = experienceList
	
	// Extract education
	resume.Education, _ = p.ExtractEducation(ctx, text)
	
	// Extract summary
	resume.Summary = p.extractSummary(text)
	
	// Extract certifications
	resume.Certifications = p.extractCertifications(text)
	
	// Extract languages
	resume.Languages = p.extractLanguages(text)
	
	return resume, nil
}