package ollama

import (
	"context"
	"fmt"
	"github.com/aqua777/ai-flow/llm/models"
)

type OllamaEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type OllamaEmbeddingResponse struct {
	Model string `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`
}

func (o *Client) Embeddings(ctx context.Context, cr *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
	req := OllamaEmbeddingRequest{
		Model: cr.Model,
		Input: cr.Content,
	}
	var resp OllamaEmbeddingResponse
	err := o.client.Post(ctx, "/api/embed", req, &resp, nil)
	if err != nil {
		return nil, err
	} else if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings found in the response")
	}
	result := &models.EmbeddingsResponse{
		Embeddings: resp.Embeddings[0],
	}
	return result, nil
}
