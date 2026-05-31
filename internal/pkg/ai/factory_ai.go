package ai

import (
	"fmt"
	"os"
)

type ProviderFactory struct {
	config *Config
}

func NewProviderFactory(config *Config) *ProviderFactory {
	return &ProviderFactory{config: config}
}

func (f *ProviderFactory) GetProvider() (AIProvider, error) {
	// Check environment variable override
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = f.config.Provider
	}
	
	switch provider {
	case "openai":
		return NewOpenAIProvider(f.config)
	case "anthropic":
		return NewAnthropicProvider(f.config)
	case "gemini":
		return NewGeminiProvider(f.config)
	case "local":
		return NewLocalLLMProvider(f.config)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
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
	
	return providers
}