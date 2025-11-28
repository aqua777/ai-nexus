package ollama

import (
	"context"

	"github.com/aqua777/ai-flow/http"
	"github.com/aqua777/ai-flow/llm/models"
)

// const (
// 	OLLAMA_HOST_ENV_VAR = "OLLAMA_HOST"
// 	OLLAMA_URL_ENV_VAR  = "OLLAMA_URL"
// 	DEAFULT_OLLAMA_URL  = "http://localhost:11434"
// )

// func getOllamaUrl() string {
// 	url := os.Getenv(OLLAMA_URL_ENV_VAR)
// 	if url == "" {
// 		host := os.Getenv(OLLAMA_HOST_ENV_VAR)
// 		if host == "" {
// 			return DEAFULT_OLLAMA_URL
// 		}
// 		url = "http://" + host
// 	}
// 	return url
// }

type Client struct {
	config *models.LLMConfig
	client *http.JsonClient
}

func (o *Client) ListModels(ctx context.Context) ([]*models.Model, error) {
	var response ListModelResponse
	err := o.client.Get(ctx, "/api/tags", &response, nil)
	if err != nil {
		return nil, err
	}
	results := make([]*models.Model, len(response.Models))
	for idx, model := range response.Models {
		results[idx] = &models.Model{
			ID:    model.Name,
			Name:  model.Name,
			Model: model.Model,
		}
	}
	return results, nil
}

func NewClient(optionalConfig ...*models.LLMConfig) (*Client, error) {
	config := models.OptionalConfig(optionalConfig).GetConfig(models.OLLAMA)
	client, err := http.NewJsonClient(config.Url)
	if err != nil {
		return nil, err
	}
	return &Client{
		config: config,
		client: client,
	}, nil
}
