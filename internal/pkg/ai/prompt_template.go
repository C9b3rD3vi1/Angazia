package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type PromptTemplates struct{}

func NewPromptTemplates() *PromptTemplates {
	return &PromptTemplates{}
}

func (p *PromptTemplates) BuildMatchAnalysisPrompt(job JobDescription, candidate CandidateProfile) string {
	var buf bytes.Buffer
	
	buf.WriteString(`You are an expert hiring assistant for the Kenyan tech market. Analyze how well this candidate matches the job position.

## JOB POSITION:
Title: ` + job.Title + `
Company Industry: ` + job.Industry + `
Experience Level Required: ` + job.ExperienceLevel + ` (` + fmt.Sprintf("%d", job.MinExperience) + `-` + fmt.Sprintf("%d", job.MaxExperience) + ` years)
Employment Type: ` + job.EmploymentType + `
Location: ` + job.Location + ` (Remote: ` + fmt.Sprintf("%t", job.IsRemote) + `)

Required Skills: ` + strings.Join(job.RequiredSkills, ", ") + `
Nice-to-Have Skills: ` + strings.Join(job.NiceToHaveSkills, ", ") + `

Job Description Summary: ` + truncateString(job.Description, 500) + `

## CANDIDATE PROFILE:
Name: ` + candidate.FullName + `
Headline: ` + candidate.Headline + `
Experience Level: ` + candidate.ExperienceLevel + ` (` + fmt.Sprintf("%d", candidate.YearsOfExperience) + ` years)
Location: ` + candidate.Location + ` (Remote Only: ` + fmt.Sprintf("%t", candidate.IsRemoteOnly) + `)

Skills: ` + strings.Join(candidate.Skills, ", ") + `
Bio: ` + truncateString(candidate.Bio, 300) + `

## GITHUB ACTIVITY (if available):
Public Repos: ` + fmt.Sprintf("%d", candidate.GithubActivity.PublicRepos) + `
Total Commits (last year): ` + fmt.Sprintf("%d", candidate.GithubActivity.TotalCommits) + `
Followers: ` + fmt.Sprintf("%d", candidate.GithubActivity.Followers) + `
Top Languages: ` + strings.Join(candidate.GithubActivity.TopLanguages, ", ") + `
Contribution Streak: ` + fmt.Sprintf("%d", candidate.GithubActivity.ContributionStreak) + ` days

Analyze and provide a JSON response with the following structure:
{
  "overall_score": 0-100,
  "skills_score": 0-100,
  "experience_score": 0-100,
  "culture_score": 0-100,
  "location_score": 0-100,
  "matching_skills": ["skill1", "skill2"],
  "missing_skills": ["skill1", "skill2"],
  "strong_points": ["point1", "point2"],
  "weak_points": ["point1", "point2"],
  "summary": "Brief 2-3 sentence summary of match quality",
  "recommendation": "hire | interview | consider | reject",
  "interview_tips": ["tip1", "tip2", "tip3"]
}

Be honest and specific. Consider Kenyan tech market context.`)

	return buf.String()
}

func (p *PromptTemplates) BuildCoverLetterPrompt(job JobDescription, candidate CandidateProfile) string {
	var buf bytes.Buffer
	
	buf.WriteString(`Write a professional, personalized cover letter for the following job application in Kenya.

## JOB DETAILS:
Company: ` + job.Industry + ` industry
Position: ` + job.Title + `
Required Skills: ` + strings.Join(job.RequiredSkills, ", ") + `
Job Description: ` + truncateString(job.Description, 400) + `

## CANDIDATE DETAILS:
Name: ` + candidate.FullName + `
Current Role/Headline: ` + candidate.Headline + `
Experience: ` + fmt.Sprintf("%d", candidate.YearsOfExperience) + ` years
Key Skills: ` + strings.Join(candidate.Skills, ", ") + `
Bio/Background: ` + truncateString(candidate.Bio, 200) + `

## GITHUB HIGHLIGHTS (if available):
- ` + fmt.Sprintf("%d", candidate.GithubActivity.TotalCommits) + ` commits in the last year
- Active in: ` + strings.Join(candidate.GithubActivity.TopLanguages, ", ") + `

Write a compelling cover letter that:
1. Shows genuine interest in the role and company
2. Highlights relevant skills and experience
3. Mentions specific GitHub contributions (if applicable)
4. Is tailored to the Kenyan tech job market
5. Is professional but not overly formal
6. Length: 250-350 words

Return ONLY the cover letter text, no additional commentary.`)

	return buf.String()
}

func (p *PromptTemplates) BuildSkillsGapPrompt(job JobDescription, candidate CandidateProfile) string {
	var buf bytes.Buffer
	
	buf.WriteString(`You are a technical career coach for the Kenyan tech industry. Analyze the skills gap between this candidate and job requirement.

## JOB REQUIREMENTS:
Title: ` + job.Title + `
Required Skills: ` + strings.Join(job.RequiredSkills, ", ") + `
Nice-to-Have Skills: ` + strings.Join(job.NiceToHaveSkills, ", ") + `
Experience Level: ` + job.ExperienceLevel + `

## CANDIDATE SKILLS:
Current Skills: ` + strings.Join(candidate.Skills, ", ") + `
Years Experience: ` + fmt.Sprintf("%d", candidate.YearsOfExperience) + `

Provide a detailed skills gap analysis as JSON:
{
  "missing_skills": [
    {
      "skill_name": "Skill Name",
      "importance": "critical|important|nice-to-have",
      "description": "Why this skill matters",
      "learning_resources": ["resource1", "resource2"]
    }
  ],
  "recommended_courses": [
    {
      "name": "Course Name",
      "platform": "Coursera|Udemy|Pluralsight|YouTube",
      "url": "https://...",
      "duration": "X weeks/months",
      "difficulty": "beginner|intermediate|advanced"
    }
  ],
  "improvement_plan": "Step-by-step plan to acquire missing skills (2-3 paragraphs)",
  "estimated_time_to_fill": "X-X months",
  "transferable_skills": ["skill1", "skill2"],
  "priority_level": "high|medium|low"
}

Consider affordable/free learning resources available in Kenya.`)

	return buf.String()
}

func (p *PromptTemplates) BuildInterviewQuestionsPrompt(job JobDescription) string {
	var buf bytes.Buffer
	
	buf.WriteString(`Generate 8-10 interview questions for a ` + job.Title + ` position in the Kenyan tech market.

## JOB DETAILS:
Title: ` + job.Title + `
Required Skills: ` + strings.Join(job.RequiredSkills, ", ") + `
Experience Level: ` + job.ExperienceLevel + `
Industry: ` + job.Industry + `

Requirements: ` + truncateString(job.Requirements, 300) + `

Generate a mix of:
- Technical questions specific to required skills
- Behavioral questions relevant to Kenyan workplace culture
- Problem-solving scenarios
- System design questions (if senior level)
- Cultural fit questions

Return as JSON array of strings.`)

	return buf.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (p *PromptTemplates) ParseMatchAnalysisResponse(response string) (*MatchAnalysis, error) {
	// Extract JSON from response
	jsonStr := extractJSON(response)
	
	var analysis MatchAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse match analysis: %w", err)
	}
	
	return &analysis, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 {
		return s
	}
	return s[start : end+1]
}