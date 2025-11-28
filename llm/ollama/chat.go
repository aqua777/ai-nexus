package ollama

import (
	"context"
	"time"

	"github.com/aqua777/ai-nexus/llm/models"
	"github.com/aqua777/ai-nexus/llm/thinking"
)

type OllamaChatCompletionRequest struct {
	Model    string                 `json:"model"`
	Messages []*models.Message      `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Context  []int                  `json:"context,omitempty"`
}

type OllamaChatCompletionResponse struct {
	Model              string          `json:"model"`
	CreatedAt          time.Time       `json:"created_at"`
	Message            *models.Message `json:"message,omitempty"`
	Response           string          `json:"response,omitempty"`
	Done               bool            `json:"done"`
	DoneReason         string          `json:"done_reason,omitempty"`
	Context            []int           `json:"context,omitempty"`
	TotalDuration      int64           `json:"total_duration,omitempty"`
	LoadDuration       int64           `json:"load_duration,omitempty"`
	PromptEvalCount    int             `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64           `json:"prompt_eval_duration,omitempty"`
	EvalCount          int             `json:"eval_count,omitempty"`
	EvalDuration       int64           `json:"eval_duration,omitempty"`
}

func (o *Client) Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error) {
	req := OllamaChatCompletionRequest{
		Model:    r.Model,
		Messages: r.Messages,
		Stream:   r.Stream,
		Options:  r.Options.ToMap(),
	}
	resp := new(OllamaChatCompletionResponse)
	err := o.client.Post(ctx, "/api/chat", req, resp, nil)
	if err != nil {
		return nil, err
	}
	content, thinking := thinking.ProcessContent(resp.Message.Content)
	return &models.ChatResponse{
		Content:   content,
		Reasoning: thinking,
		Metadata: &models.ChatResponseMetadata{
			PromptTokens:     resp.PromptEvalCount,
			CompletionTokens: resp.EvalCount,
			TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
		},
	}, nil
}
