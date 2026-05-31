package ai

import (
	"context"
	"fmt"
)

type LocalLLMProvider struct {
	config  *Config
	prompts *PromptTemplates
}

func NewLocalLLMProvider(config *Config) (*LocalLLMProvider, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("Local LLM base URL is required")
	}
	
	return &LocalLLMProvider{
		config:  config,
		prompts: NewPromptTemplates(),
	}, nil
}

func (p *LocalLLMProvider) GetProviderName() string {
	return "local"
}

func (p *LocalLLMProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (p *LocalLLMProvider) GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error) {
	return nil, fmt.Errorf("Local LLM provider not fully implemented yet")
}

func (p *LocalLLMProvider) GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error) {
	return "", fmt.Errorf("Local LLM provider not fully implemented yet")
}

func (p *LocalLLMProvider) AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error) {
	return nil, fmt.Errorf("Local LLM provider not fully implemented yet")
}

func (p *LocalLLMProvider) GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error) {
	return nil, fmt.Errorf("Local LLM provider not fully implemented yet")
}