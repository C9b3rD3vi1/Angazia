package ai

import (
	"fmt"
	"os"
)

type ProviderFactory struct {
	config *Config
}

func NewProviderFactory(config *Config) *ProviderFactory {
	if config == nil {
		config = DefaultConfig()
	}
	return &ProviderFactory{config: config}
}

func (f *ProviderFactory) GetProvider() (AIProvider, error) {
	// Check environment variable override
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = f.config.Provider
	}
	
	// Get API keys from environment if not in config
	apiKey := f.config.APIKey
	baseURL := f.config.BaseURL
	
	switch provider {
	case "openai":
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		config := &Config{
			Provider:      "openai",
			APIKey:        apiKey,
			Model:         f.config.Model,
			Timeout:       f.config.Timeout,
			MaxTokens:     f.config.MaxTokens,
			Temperature:   f.config.Temperature,
			RetryAttempts: f.config.RetryAttempts,
			RetryDelay:    f.config.RetryDelay,
		}
		return NewOpenAIProvider(config)
		
	case "anthropic":
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required for Anthropic provider")
		}
		config := &Config{
			Provider:      "anthropic",
			APIKey:        apiKey,
			Model:         f.config.Model,
			Timeout:       f.config.Timeout,
			MaxTokens:     f.config.MaxTokens,
			Temperature:   f.config.Temperature,
			RetryAttempts: f.config.RetryAttempts,
			RetryDelay:    f.config.RetryDelay,
		}
		return NewAnthropicProvider(config)
		
	case "gemini":
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required for Gemini provider")
		}
		config := &Config{
			Provider:      "gemini",
			APIKey:        apiKey,
			Model:         f.config.Model,
			Timeout:       f.config.Timeout,
			MaxTokens:     f.config.MaxTokens,
			Temperature:   f.config.Temperature,
			RetryAttempts: f.config.RetryAttempts,
			RetryDelay:    f.config.RetryDelay,
		}
		return NewGeminiProvider(config)
		
	case "local":
		if baseURL == "" {
			baseURL = os.Getenv("LOCAL_LLM_URL")
		}
		if baseURL == "" {
			return nil, fmt.Errorf("LOCAL_LLM_URL environment variable is required for local provider")
		}
		config := &Config{
			Provider:      "local",
			BaseURL:       baseURL,
			Model:         f.config.Model,
			Timeout:       f.config.Timeout,
			MaxTokens:     f.config.MaxTokens,
			Temperature:   f.config.Temperature,
			RetryAttempts: f.config.RetryAttempts,
			RetryDelay:    f.config.RetryDelay,
		}
		return NewLocalLLMProvider(config)
		
	default:
		// Default to OpenAI with mock mode if no API key
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		config := &Config{
			Provider:      "openai",
			APIKey:        apiKey,
			Model:         f.config.Model,
			Timeout:       f.config.Timeout,
			MaxTokens:     f.config.MaxTokens,
			Temperature:   f.config.Temperature,
			RetryAttempts: f.config.RetryAttempts,
			RetryDelay:    f.config.RetryDelay,
		}
		return NewOpenAIProvider(config)
	}
}

// GetAvailableProviders returns list of configured providers
func (f *ProviderFactory) GetAvailableProviders() []string {
	providers := []string{}
	
	if os.Getenv("OPENAI_API_KEY") != "" {
		providers = append(providers, "openai")
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providers = append(providers, "anthropic")
	}
	if os.Getenv("GEMINI_API_KEY") != "" {
		providers = append(providers, "gemini")
	}
	if os.Getenv("LOCAL_LLM_URL") != "" {
		providers = append(providers, "local")
	}
	
	// Always add openai as fallback (will use mock mode)
	if len(providers) == 0 {
		providers = append(providers, "openai (mock mode)")
	}
	
	return providers
}

// GetDefaultProvider returns the default provider name
func (f *ProviderFactory) GetDefaultProvider() string {
	if len(f.GetAvailableProviders()) > 0 {
		return f.GetAvailableProviders()[0]
	}
	return "openai"
}