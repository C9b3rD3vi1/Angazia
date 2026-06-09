package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AnthropicProvider struct {
	config     *Config
	httpClient *http.Client
	prompts    *PromptTemplates
	apiURL     string
}

type AnthropicRequest struct {
	Model         string             `json:"model"`
	MaxTokens     int                `json:"max_tokens"`
	Temperature   float64            `json:"temperature"`
	System        string             `json:"system"`
	Messages      []AnthropicMessage `json:"messages"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Role       string `json:"role"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewAnthropicProvider(config *Config) (*AnthropicProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}
	
	if config.Model == "" {
		config.Model = "claude-3-sonnet-20240229"
	}
	
	return &AnthropicProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		prompts: NewPromptTemplates(),
		apiURL:  "https://api.anthropic.com/v1/messages",
	}, nil
}

func (p *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

func (p *AnthropicProvider) HealthCheck(ctx context.Context) error {
	_, err := p.makeRequest(ctx, []AnthropicMessage{{Role: "user", Content: "test"}})
	return err
}

func (p *AnthropicProvider) GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error) {
	startTime := time.Now()
	
	if p.prompts == nil {
		p.prompts = NewPromptTemplates()
	}
	
	prompt := p.prompts.BuildMatchAnalysisPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []AnthropicMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
	}
	
	var textContent string
	for _, content := range response.Content {
		if content.Type == "text" {
			textContent = content.Text
			break
		}
	}
	
	analysis, err := p.prompts.ParseMatchAnalysisResponse(textContent)
	if err != nil {
		return nil, err
	}
	
	analysis.AnalysisMetadata = AnalysisMetadata{
		Provider:         p.GetProviderName(),
		Model:            p.config.Model,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		AnalyzedAt:       time.Now(),
		TokensUsed:       response.Usage.InputTokens + response.Usage.OutputTokens,
	}
	
	return analysis, nil
}

func (p *AnthropicProvider) GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error) {
	if p.prompts == nil {
		p.prompts = NewPromptTemplates()
	}
	
	prompt := p.prompts.BuildCoverLetterPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []AnthropicMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate cover letter: %w", err)
	}
	
	for _, content := range response.Content {
		if content.Type == "text" {
			return content.Text, nil
		}
	}
	
	return "", fmt.Errorf("no text content in response")
}

func (p *AnthropicProvider) AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error) {
	if p.prompts == nil {
		p.prompts = NewPromptTemplates()
	}
	
	prompt := p.prompts.BuildSkillsGapPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []AnthropicMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze skills gap: %w", err)
	}
	
	var textContent string
	for _, content := range response.Content {
		if content.Type == "text" {
			textContent = content.Text
			break
		}
	}
	
	var analysis SkillsGapAnalysis
	if err := json.Unmarshal([]byte(extractJSON(textContent)), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse skills gap analysis: %w", err)
	}
	
	return &analysis, nil
}

func (p *AnthropicProvider) GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error) {
	if p.prompts == nil {
		p.prompts = NewPromptTemplates()
	}
	
	prompt := p.prompts.BuildInterviewQuestionsPrompt(job)
	
	response, err := p.makeRequest(ctx, []AnthropicMessage{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate interview questions: %w", err)
	}
	
	var textContent string
	for _, content := range response.Content {
		if content.Type == "text" {
			textContent = content.Text
			break
		}
	}
	
	var questions []string
	if err := json.Unmarshal([]byte(extractJSON(textContent)), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse interview questions: %w", err)
	}
	
	return questions, nil
}

func (p *AnthropicProvider) makeRequest(ctx context.Context, messages []AnthropicMessage) (*AnthropicResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt < p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(p.config.RetryDelay)
		}
		
		reqBody := AnthropicRequest{
			Model:       p.config.Model,
			MaxTokens:   p.config.MaxTokens,
			Temperature: p.config.Temperature,
			System:      "You are an expert hiring assistant for the Kenyan tech market. Respond only with valid JSON when requested.",
			Messages:    messages,
		}
		
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			lastErr = err
			continue
		}
		
		req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			lastErr = err
			continue
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", p.config.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		
		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		
		var anthropicResp AnthropicResponse
		if err := json.Unmarshal(body, &anthropicResp); err != nil {
			lastErr = err
			continue
		}
		
		if anthropicResp.Error != nil {
			lastErr = fmt.Errorf("Anthropic API error: %s - %s", anthropicResp.Error.Type, anthropicResp.Error.Message)
			continue
		}
		
		return &anthropicResp, nil
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", p.config.RetryAttempts, lastErr)
}