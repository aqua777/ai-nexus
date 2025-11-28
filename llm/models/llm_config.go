package models

import (
	"fmt"
	"os"
	"strings"
)

const (
	OPENAI = "openai"
	OLLAMA = "ollama"
	
	DEFAULT_OPENAI_URL_V1 = "https://api.openai.com/v1"
	DEFAULT_OLLAMA_URL = "http://localhost:11434"
)

type LLMConfig struct {
	Provider string `json:"provider"`
	Url    string `json:"url"`
	ApiKey string `json:"api_key"`
}

var providerDefaultUrls = map[string]string{
	"openai": DEFAULT_OPENAI_URL_V1,
	"ollama": DEFAULT_OLLAMA_URL,
}

func (c *LLMConfig) WithDefaults(provider string) *LLMConfig {
	// Apply defaults for URL if not set
	if c.Url == "" {
		urlEnvVar := fmt.Sprintf("%s_URL", strings.ToUpper(provider))
		c.Url = os.Getenv(urlEnvVar)
		if c.Url == "" {
			c.Url = providerDefaultUrls[provider]
		}
	}
	// Apply defaults for ApiKey if not set
	if c.ApiKey == "" {
		apiKeyEnvVar := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
		c.ApiKey = os.Getenv(apiKeyEnvVar)
	}
	return &LLMConfig{
		Provider: c.Provider,
		Url:    c.Url,
		ApiKey: c.ApiKey,
	}
}

type OptionalConfig []*LLMConfig

func (me OptionalConfig) GetConfig(provider string) *LLMConfig {
	var config *LLMConfig
	if len(me) == 1 {
		config = me[0]
	} else {
		config = &LLMConfig{}
	}
	return config.WithDefaults(provider)
}
