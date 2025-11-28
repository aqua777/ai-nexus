package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aqua777/ai-flow/llm/models"
	"github.com/aqua777/ai-flow/llm/ollama"
)

var model string = os.Getenv("OLLAMA_MODEL")

func main() {
	ctx := context.Background()
	ollama, err := ollama.NewClient()
	if err != nil {
		slog.Error("Failed to create Ollama client", "error", err)
	}
	chatCompletionRequest := &models.ChatRequest{
		Model: model,
		Messages: []*models.Message{
			{Role: models.SystemRole, Content: "You are a helpful assistant."},
			{Role: models.UserRole, Content: "What does ARR stand for?"},
		},
	}
	chatCompletionResponse, err := ollama.Chat(ctx, chatCompletionRequest)
	if err != nil {
		slog.Error("Failed to chat completion", "error", err)
	}
	resp, err := json.MarshalIndent(chatCompletionResponse, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal chat completion response", "error", err)
	}
	fmt.Println("Chat completion response", string(resp))
}
