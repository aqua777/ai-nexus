package iface

import (
	"context"

	"github.com/aqua777/ai-nexus/llm/models"
)

type LLM interface {
	ListModels(ctx context.Context) ([]*models.Model, error)
	Generate(ctx context.Context, r *models.GenerateRequest) (*models.GenerateResponse, error)
	Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error)
	Embeddings(ctx context.Context, cr *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error)
}
