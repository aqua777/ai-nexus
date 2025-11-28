package main


import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aqua777/ai-flow/llm/ollama"
)

func main() {
	ctx := context.Background()
	ollama, err := ollama.NewClient(nil)
	if err != nil {
		slog.Error("Failed to create Ollama client", "error", err)
	}
	models, err := ollama.ListModels(ctx)
	if err != nil {
		slog.Error("Failed to list models", "error", err)
	}
	for _, model := range models {
		fmt.Println("Model", model)
	}
}
