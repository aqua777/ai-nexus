package llm

import (
	"context"
	"errors"

	"github.com/aqua777/ai-flow/llm/iface"
	"github.com/aqua777/ai-flow/llm/models"
)

type (
	LLM                    = iface.LLM
	Model                  = models.Model
	LLMConfig              = models.LLMConfig
	ChatCompletionRequest  = models.ChatRequest
	ChatCompletionResponse = models.ChatResponse
)

func New(ctx context.Context, config *LLMConfig) (LLM, error) {
	return nil, errors.New("not implemented")
}
