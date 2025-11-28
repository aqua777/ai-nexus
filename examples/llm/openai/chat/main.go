package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "github.com/aqua777/ai-nexus/dotenv"
	"github.com/aqua777/ai-nexus/llm/models"
	"github.com/aqua777/ai-nexus/llm/openai"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()
	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		for _, env := range os.Environ() {
			fmt.Println(env)
		}
	}

	ctx := context.Background()
	openaiClient, err := openai.NewClient()
	if err != nil {
		log.Fatalf("Failed to create OpenAI client: %v", err)
	}

	chatCompletionRequest := &models.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []*models.Message{
			{Role: models.SystemRole, Content: "You are a helpful assistant."},
			{Role: models.UserRole, Content: "What does ARR stand for?"},
		},
	}

	chatCompletionResponse, err := openaiClient.Chat(ctx, chatCompletionRequest)
	if err != nil {
		log.Fatalf("Failed to chat completion: %v", err)
	}

	fmt.Println("Response:", chatCompletionResponse.Content)
}
