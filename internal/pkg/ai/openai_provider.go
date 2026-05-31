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

type OpenAIProvider struct {
	config      *Config
	httpClient  *http.Client
	prompts     *PromptTemplates
	apiURL      string
}

type OpenAIChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewOpenAIProvider(config *Config) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	
	if config.Model == "" {
		config.Model = "gpt-4-turbo-preview"
	}
	
	return &OpenAIProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		prompts: NewPromptTemplates(),
		apiURL:  "https://api.openai.com/v1/chat/completions",
	}, nil
}

func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	_, err := p.makeRequest(ctx, []Message{{Role: "user", Content: "test"}})
	return err
}

func (p *OpenAIProvider) GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error) {
	startTime := time.Now()
	
	prompt := p.prompts.BuildMatchAnalysisPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []Message{
		{Role: "system", Content: "You are an expert hiring assistant. Respond only with valid JSON."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
	}
	
	analysis, err := p.prompts.ParseMatchAnalysisResponse(response.Content)
	if err != nil {
		return nil, err
	}
	
	analysis.AnalysisMetadata = AnalysisMetadata{
		Provider:        p.GetProviderName(),
		Model:           p.config.Model,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		AnalyzedAt:      time.Now(),
	}
	
	return analysis, nil
}

func (p *OpenAIProvider) GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error) {
	prompt := p.prompts.BuildCoverLetterPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []Message{
		{Role: "system", Content: "You are a professional cover letter writer. Write compelling, personalized cover letters."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate cover letter: %w", err)
	}
	
	return response.Content, nil
}

func (p *OpenAIProvider) AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error) {
	prompt := p.prompts.BuildSkillsGapPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []Message{
		{Role: "system", Content: "You are a technical career coach. Respond only with valid JSON."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze skills gap: %w", err)
	}
	
	var analysis SkillsGapAnalysis
	if err := json.Unmarshal([]byte(response.Content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse skills gap analysis: %w", err)
	}
	
	return &analysis, nil
}

func (p *OpenAIProvider) GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error) {
	prompt := p.prompts.BuildInterviewQuestionsPrompt(job)
	
	response, err := p.makeRequest(ctx, []Message{
		{Role: "system", Content: "You are an experienced technical interviewer. Respond only with valid JSON array of strings."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate interview questions: %w", err)
	}
	
	var questions []string
	if err := json.Unmarshal([]byte(response.Content), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse interview questions: %w", err)
	}
	
	return questions, nil
}

func (p *OpenAIProvider) makeRequest(ctx context.Context, messages []Message) (*Message, error) {
	var lastErr error
	
	for attempt := 0; attempt < p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(p.config.RetryDelay)
		}
		
		reqBody := OpenAIChatRequest{
			Model:       p.config.Model,
			Messages:    messages,
			Temperature: p.config.Temperature,
			MaxTokens:   p.config.MaxTokens,
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
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
		
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
		
		var chatResp OpenAIChatResponse
		if err := json.Unmarshal(body, &chatResp); err != nil {
			lastErr = err
			continue
		}
		
		if chatResp.Error != nil {
			lastErr = fmt.Errorf("OpenAI API error: %s", chatResp.Error.Message)
			continue
		}
		
		if len(chatResp.Choices) == 0 {
			lastErr = fmt.Errorf("no response from OpenAI")
			continue
		}
		
		return &chatResp.Choices[0].Message, nil
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", p.config.RetryAttempts, lastErr)
}