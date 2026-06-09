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

type GeminiProvider struct {
	config     *Config
	httpClient *http.Client
	prompts    *PromptTemplates
	apiURL     string
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewGeminiProvider(config *Config) (*GeminiProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	
	if config.Model == "" {
		config.Model = "gemini-1.5-pro"
	}
	
	return &GeminiProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		prompts: NewPromptTemplates(),
		apiURL:  fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/%s:generateContent?key=%s", config.Model, config.APIKey),
	}, nil
}

func (p *GeminiProvider) GetProviderName() string {
	return "gemini"
}

func (p *GeminiProvider) HealthCheck(ctx context.Context) error {
	_, err := p.makeRequest(ctx, []GeminiContent{
		{Parts: []GeminiPart{{Text: "test"}}},
	})
	return err
}

func (p *GeminiProvider) GenerateMatchAnalysis(ctx context.Context, job JobDescription, candidate CandidateProfile) (*MatchAnalysis, error) {
	startTime := time.Now()
	
	prompt := p.prompts.BuildMatchAnalysisPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []GeminiContent{
		{Parts: []GeminiPart{{Text: prompt}}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
	}
	
	var textContent string
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		textContent = response.Candidates[0].Content.Parts[0].Text
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
		TokensUsed:       response.UsageMetadata.TotalTokenCount,
	}
	
	return analysis, nil
}

func (p *GeminiProvider) GenerateCoverLetter(ctx context.Context, job JobDescription, candidate CandidateProfile) (string, error) {
	prompt := p.prompts.BuildCoverLetterPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []GeminiContent{
		{Parts: []GeminiPart{{Text: prompt}}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate cover letter: %w", err)
	}
	
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		return response.Candidates[0].Content.Parts[0].Text, nil
	}
	
	return "", fmt.Errorf("no content in response")
}

func (p *GeminiProvider) AnalyzeSkillsGap(ctx context.Context, job JobDescription, candidate CandidateProfile) (*SkillsGapAnalysis, error) {
	prompt := p.prompts.BuildSkillsGapPrompt(job, candidate)
	
	response, err := p.makeRequest(ctx, []GeminiContent{
		{Parts: []GeminiPart{{Text: prompt}}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze skills gap: %w", err)
	}
	
	var textContent string
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		textContent = response.Candidates[0].Content.Parts[0].Text
	}
	
	var analysis SkillsGapAnalysis
	if err := json.Unmarshal([]byte(extractJSON(textContent)), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse skills gap analysis: %w", err)
	}
	
	return &analysis, nil
}

func (p *GeminiProvider) GenerateInterviewQuestions(ctx context.Context, job JobDescription) ([]string, error) {
	prompt := p.prompts.BuildInterviewQuestionsPrompt(job)
	
	response, err := p.makeRequest(ctx, []GeminiContent{
		{Parts: []GeminiPart{{Text: prompt}}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate interview questions: %w", err)
	}
	
	var textContent string
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		textContent = response.Candidates[0].Content.Parts[0].Text
	}
	
	var questions []string
	if err := json.Unmarshal([]byte(extractJSON(textContent)), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse interview questions: %w", err)
	}
	
	return questions, nil
}

func (p *GeminiProvider) makeRequest(ctx context.Context, contents []GeminiContent) (*GeminiResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt < p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(p.config.RetryDelay)
		}
		
		reqBody := GeminiRequest{
			Contents: contents,
			GenerationConfig: GeminiGenerationConfig{
				Temperature:     p.config.Temperature,
				MaxOutputTokens: p.config.MaxTokens,
			},
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
		
		var geminiResp GeminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			lastErr = err
			continue
		}
		
		if geminiResp.Error != nil {
			lastErr = fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
			continue
		}
		
		return &geminiResp, nil
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", p.config.RetryAttempts, lastErr)
}