package models

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptionalConfigTestSuite struct {
	suite.Suite
	originalEnvVars      map[string]string
}

func TestOptionalConfigTestSuite(t *testing.T) {
	suite.Run(t, new(OptionalConfigTestSuite))
}

func (s *OptionalConfigTestSuite) SetupTest() {
	s.originalEnvVars = make(map[string]string)
	for _, provider := range []string{OPENAI, OLLAMA} {
		for _, envVar := range []string{"URL", "API_KEY"} {
			envVarName := fmt.Sprintf("%s_%s", strings.ToUpper(provider), strings.ToUpper(envVar))
			// Save original environment variables
			s.originalEnvVars[envVarName] = os.Getenv(envVarName)
			// Clear environment variables for clean test state
			os.Unsetenv(envVarName)
		}
	}
}

func (s *OptionalConfigTestSuite) TearDownTest() {
	// Restore original environment variables
	for envVarName, value := range s.originalEnvVars {
		os.Setenv(envVarName, value)
	}
}

func (s *OptionalConfigTestSuite) TestAll() {
	testData := []struct {
		name           string
		provider       string
		envUrl         string
		envApiKey      string
		configUrl      string
		configApiKey   string
		expectedUrl    string
		expectedApiKey string
	}{
		// OpenAI
		{name: "OpenAI_NeitherDefined", provider: OPENAI, envUrl: "", envApiKey: "", configUrl: "", configApiKey: "", expectedUrl: "https://api.openai.com/v1", expectedApiKey: ""},
		{name: "OpenAI_UrlDefined_ApiKeyNot", provider: OPENAI, envUrl: "", envApiKey: "", configUrl: "https://custom.openai.com/v1", configApiKey: "", expectedUrl: "https://custom.openai.com/v1", expectedApiKey: ""},
		{name: "OpenAI_EnvUrlDefined_ApiKeyNot", provider: OPENAI, envUrl: "https://custom.openai.com/v1", envApiKey: "", configUrl: "", configApiKey: "", expectedUrl: "https://custom.openai.com/v1", expectedApiKey: ""},
		{name: "OpenAI_ApiKeyDefined_UrlNot", provider: OPENAI, envUrl: "", envApiKey: "", configUrl: "", configApiKey: "my-key", expectedUrl: "https://api.openai.com/v1", expectedApiKey: "my-key"},
		{name: "OpenAI_EnvApiKeyDefined_UrlNot", provider: OPENAI, envUrl: "", envApiKey: "my-key", configUrl: "", configApiKey: "", expectedUrl: "https://api.openai.com/v1", expectedApiKey: "my-key"},
		{name: "OpenAI_ApiKeyAndUrlDefined", provider: OPENAI, envUrl: "", envApiKey: "", configUrl: "https://custom.openai.com/v1", configApiKey: "my-key", expectedUrl: "https://custom.openai.com/v1", expectedApiKey: "my-key"},
		{name: "OpenAI_EnvApiKeyAndUrlDefined", provider: OPENAI, envUrl: "https://custom.openai.com/v1", envApiKey: "my-key", configUrl: "", configApiKey: "", expectedUrl: "https://custom.openai.com/v1", expectedApiKey: "my-key"},
		// Ollama
		{name: "Ollama_NeitherDefined", provider: OLLAMA, envUrl: "", envApiKey: "", configUrl: "", configApiKey: "", expectedUrl: "http://localhost:11434", expectedApiKey: ""},
		{name: "Ollama_UrlDefined_ApiKeyNot", provider: OLLAMA, envUrl: "", envApiKey: "", configUrl: "https://custom.ollama.com/v1", configApiKey: "", expectedUrl: "https://custom.ollama.com/v1", expectedApiKey: ""},
		{name: "Ollama_EnvUrlDefined_ApiKeyNot", provider: OLLAMA, envUrl: "https://custom.ollama.com/v1", envApiKey: "", configUrl: "", configApiKey: "", expectedUrl: "https://custom.ollama.com/v1", expectedApiKey: ""},
		{name: "Ollama_ApiKeyDefined_UrlNot", provider: OLLAMA, envUrl: "", envApiKey: "", configUrl: "", configApiKey: "my-key", expectedUrl: "http://localhost:11434", expectedApiKey: "my-key"},
		{name: "Ollama_EnvApiKeyDefined_UrlNot", provider: OLLAMA, envUrl: "", envApiKey: "my-key", configUrl: "", configApiKey: "", expectedUrl: "http://localhost:11434", expectedApiKey: "my-key"},
		{name: "Ollama_ApiKeyAndUrlDefined", provider: OLLAMA, envUrl: "", envApiKey: "", configUrl: "https://custom.ollama.com/v1", configApiKey: "my-key", expectedUrl: "https://custom.ollama.com/v1", expectedApiKey: "my-key"},
		{name: "Ollama_EnvApiKeyAndUrlDefined", provider: OLLAMA, envUrl: "https://custom.ollama.com/v1", envApiKey: "my-key", configUrl: "", configApiKey: "", expectedUrl: "https://custom.ollama.com/v1", expectedApiKey: "my-key"},
		// Unknown provider
		{name: "UnknownProvider", provider: "unknown", envUrl: "", envApiKey: "", configUrl: "", configApiKey: "", expectedUrl: "", expectedApiKey: ""},
	}
	for _, t := range testData {
		s.Run(t.name, func() {
			os.Setenv(fmt.Sprintf("%s_URL", strings.ToUpper(t.provider)), t.envUrl)
			os.Setenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(t.provider)), t.envApiKey)
			config := OptionalConfig([]*LLMConfig{
				{Url: t.configUrl, ApiKey: t.configApiKey},
			}).GetConfig(t.provider)
			s.Equal(t.expectedUrl, config.Url)
			s.Equal(t.expectedApiKey, config.ApiKey)
		})
	}
}
