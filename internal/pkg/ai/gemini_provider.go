package ai

import (
	"context"
	"fmt"
)

type GeminiProvider struct {
	config  *Config
	prompts *PromptTemplates
}

func NewGeminiProvider(config *Config) (*GeminiProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	
	return &GeminiProvider{
		config:  config,
		prompts: NewPromptTemplates(),
	}, nil
}

func (p *GeminiProvider) GetProviderName() string {
	return "gemini"
}

func (p *GeminiProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (p *GeminiProvider) GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error) {
	return nil, fmt.Errorf("Gemini provider not fully implemented yet")
}

func (p *GeminiProvider) GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error) {
	return "", fmt.Errorf("Gemini provider not fully implemented yet")
}

func (p *GeminiProvider) AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error) {
	return nil, fmt.Errorf("Gemini provider not fully implemented yet")
}

func (p *GeminiProvider) GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error) {
	return nil, fmt.Errorf("Gemini provider not fully implemented yet")
}