package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/aqua777/ai-nexus/llm/models"
	"github.com/aqua777/ai-nexus/llm/ollama"
)

var model string = os.Getenv("OLLAMA_MODEL")

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()
	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	ctx := context.Background()
	ollama, err := ollama.NewClient()
	if err != nil {
		slog.Error("Failed to create Ollama client", "error", err)
	}
	generateRequest := &models.GenerateRequest{
		Model:  model,
		Prompt: "Why is the sky blue?",
	}
	generateResponse, err := ollama.Generate(ctx, generateRequest)
	if err != nil {
		slog.Error("Failed to generate", "error", err)
	}
	resp, err := json.MarshalIndent(generateResponse, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal chat completion response", "error", err)
	}
	fmt.Println("Chat completion response", string(resp))
}
