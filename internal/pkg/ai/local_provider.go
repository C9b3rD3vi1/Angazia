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

type LocalLLMProvider struct {
	config     *Config
	httpClient *http.Client
	prompts    *PromptTemplates
}

type LocalLLMRequest struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Stream      bool    `json:"stream"`
}

type LocalLLMResponse struct {
	Text      string `json:"text"`
	Response  string `json:"response"`
	Content   string `json:"content"`
	Choices   []struct {
		Text   string `json:"text"`
		Index  int    `json:"index"`
		Finish string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewLocalLLMProvider(config *Config) (*LocalLLMProvider, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434/api/generate"
	}
	
	return &LocalLLMProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		prompts: NewPromptTemplates(),
	}, nil
}

func (p *LocalLLMProvider) GetProviderName() string {
	return "local"
}

func (p *LocalLLMProvider) HealthCheck(ctx context.Context) error {
	_, err := p.makeRequest(ctx, "test")
	return err
}

func (p *LocalLLMProvider) GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error) {
	startTime := time.Now()
	
	prompt := p.prompts.BuildMatchAnalysisPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
	}
	
	analysis, err := p.prompts.ParseMatchAnalysisResponse(response)
	if err != nil {
		return nil, err
	}
	
	analysis.AnalysisMetadata = AnalysisMetadata{
		Provider:         p.GetProviderName(),
		Model:            p.config.Model,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
		AnalyzedAt:       time.Now(),
	}
	
	return analysis, nil
}

func (p *LocalLLMProvider) GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error) {
	prompt := p.prompts.BuildCoverLetterPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate cover letter: %w", err)
	}
	
	return response, nil
}

func (p *LocalLLMProvider) AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error) {
	prompt := p.prompts.BuildSkillsGapPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze skills gap: %w", err)
	}
	
	var analysis SkillsGapAnalysis
	if err := json.Unmarshal([]byte(extractJSON(response)), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse skills gap analysis: %w", err)
	}
	
	return &analysis, nil
}

func (p *LocalLLMProvider) GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error) {
	prompt := p.prompts.BuildInterviewQuestionsPrompt(job)
	
	response, err := p.makeRequest(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate interview questions: %w", err)
	}
	
	var questions []string
	if err := json.Unmarshal([]byte(extractJSON(response)), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse interview questions: %w", err)
	}
	
	return questions, nil
}

func (p *LocalLLMProvider) makeRequest(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	
	for attempt := 0; attempt < p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(p.config.RetryDelay)
		}
		
		reqBody := LocalLLMRequest{
			Prompt:      prompt,
			MaxTokens:   p.config.MaxTokens,
			Temperature: p.config.Temperature,
			Stream:      false,
		}
		
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			lastErr = err
			continue
		}
		
		url := p.config.BaseURL
		if url == "" {
			url = "http://localhost:8080/v1/completions"
		}
		
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			lastErr = err
			continue
		}
		
		req.Header.Set("Content-Type", "application/json")
		
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
		
		var llmResp LocalLLMResponse
		if err := json.Unmarshal(body, &llmResp); err != nil {
			// Try to parse as simple response
			if len(body) > 0 {
				return string(body), nil
			}
			lastErr = err
			continue
		}
		
		if llmResp.Error != nil {
			lastErr = fmt.Errorf("Local LLM error: %s", llmResp.Error.Message)
			continue
		}
		
		// Extract text from various possible response formats
		if llmResp.Text != "" {
			return llmResp.Text, nil
		}
		if llmResp.Response != "" {
			return llmResp.Response, nil
		}
		if llmResp.Content != "" {
			return llmResp.Content, nil
		}
		if len(llmResp.Choices) > 0 && llmResp.Choices[0].Text != "" {
			return llmResp.Choices[0].Text, nil
		}
		
		return string(body), nil
	}
	
	return "", fmt.Errorf("failed after %d attempts: %w", p.config.RetryAttempts, lastErr)
}