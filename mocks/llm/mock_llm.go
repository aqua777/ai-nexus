package llm

import (
	"context"

	llm_iface "github.com/aqua777/ai-nexus/llm/iface"
	llm_models "github.com/aqua777/ai-nexus/llm/models"
)

type MockLLM struct{}

// Ensure MockLLM implements LLM interface
var _ llm_iface.LLM = (*MockLLM)(nil)

func (m *MockLLM) ListModels(ctx context.Context) ([]*llm_models.Model, error) {
	return nil, nil
}

func (m *MockLLM) Generate(ctx context.Context, r *llm_models.GenerateRequest) (*llm_models.GenerateResponse, error) {
	return nil, nil
}

func (m *MockLLM) Chat(ctx context.Context, r *llm_models.ChatRequest, stream ...func(chunk []byte) error) (*llm_models.ChatResponse, error) {
	return nil, nil
}

func (m *MockLLM) Embeddings(ctx context.Context, cr *llm_models.EmbeddingsRequest) (*llm_models.EmbeddingsResponse, error) {
	return &llm_models.EmbeddingsResponse{
		Embeddings: []float32{1.0, 0.0, 0.0},
	}, nil
}
