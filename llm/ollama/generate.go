package ollama

import (
	"context"
	"log/slog"
	"time"

	"github.com/aqua777/ai-nexus/llm/models"
)

type OllamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaGenerateResponse struct {
	Response   string    `json:"response"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	Done       bool      `json:"done"`
	DoneReason string    `json:"done_reason"`
	// Context []int `json:"context"`
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

func (o *Client) Generate(ctx context.Context, r *models.GenerateRequest) (*models.GenerateResponse, error) {
	req := OllamaGenerateRequest{
		Model:   r.Model,
		Prompt:  r.Prompt,
		Stream:  r.Stream,
		Options: r.Options.ToMap(),
	}
	slog.Info("Generate request", "request", req)
	var resp OllamaGenerateResponse
	err := o.client.Post(ctx, "/api/generate", req, &resp, nil)
	if err != nil {
		return nil, err
	}
	slog.Info("Generate response", "response", resp)
	return &models.GenerateResponse{
		Text:             resp.Response,
		Model:            resp.Model,
		CreatedAt:        resp.CreatedAt,
		PromptTokens:     resp.PromptEvalCount,
		CompletionTokens: resp.EvalCount,
		TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
	}, nil
}
