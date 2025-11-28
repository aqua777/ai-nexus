package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aqua777/ai-nexus/llm/ollama"
	"github.com/aqua777/ai-nexus/llm/openai"
	"github.com/aqua777/ai-nexus/rag/v2"
	"github.com/aqua777/ai-nexus/vectordb/v1/chromem"
)

var (
	openaiClient *openai.Client
	ollamaClient *ollama.Client
)

func main() {
	ctx := context.Background()

	// // 1. Setup Custom Clients (Ollama)
	// // We'll use the OpenAI client compatibility for Ollama
	// // Ollama usually runs on http://localhost:11434/v1
	// ollamaConfig := openai.DefaultConfig("ollama") // API Key doesn't matter for Ollama
	// ollamaConfig.BaseURL = "http://host.docker.internal:11434/v1"

	// ollamaClient := openai.NewClientWithConfig(ollamaConfig)

	// 2. Create Custom Implementations
	// Wrap the sashabaranov/go-openai client with our unified llm adapter.
	// The adapter implements iface.LLM which covers both embedding and chat.
	// openaiClient := llm_openai.NewClientWithOpenAIClient(ollamaClient)
	var err error
	openaiClient, err = openai.NewClient()
	if err != nil {
		log.Fatalf("Failed to create custom client: %v", err)
	}

	ollamaClient, err = ollama.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// 3. Initialize RAG System without OpenAI Key
	config := &rag.RAGConfig{
		ChunkSize:      512,
		ChunkOverlap:   50,
		PersistPath:    "./.chromem-db",
		CollectionName: "demo-docs",
		// Specify the models to be used by the client
		EmbeddingModel: "mxbai-embed-large",
		LLMModel:       "jan-v1:q6_k",
		FileExtensions: []string{".txt", ".md", ".not-txt"},
	}

	// Vector Store
	// ChromemStore implements VectorStore interface
	vectorStore, err := chromem.NewChromemStore(config.PersistPath, config.CollectionName)
	if err != nil {
		log.Fatalf("failed to create vector store: %w", err)
	}

	// Note: We don't pass OpenAIKey here because we are injecting dependencies
	sys, err := rag.NewRAGSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Inject Dependencies
	// Since our customClient implements the unified interface, we pass it to both.
	sys.WithEmbedding(ollamaClient).WithLLM(ollamaClient).WithVectorStore(vectorStore)

	// Define Callbacks
	sys.WithOnIngestStarted(func(totalDocs int) {
		fmt.Printf("--- Ingestion Started: %d documents ---\n", totalDocs)
	}).WithOnIngestProgress(func(p rag.IngestProgress) {
		fmt.Printf("Progress: Doc %d/%d - Chunk %d/%d - %s\n",
			p.CurrentDocumentIndex+1, p.TotalDocuments,
			p.CurrentChunkIndex+1, p.TotalChunksInDoc,
			p.Message)
	}).WithOnIngestCompleted(func() {
		fmt.Println("--- Ingestion Completed Successfully ---")
	}).WithOnIngestError(func(err error) {
		fmt.Printf("--- Ingestion Error: %v ---\n", err)
	})

	// 5. Ingest Data
	// Ingest a directory
	fmt.Println("Ingesting directory...")
	err = sys.IngestDirectory(ctx, "./data")
	if err != nil {
		log.Fatalf("Failed to ingest directory: %v", err)
	}

	// Ingest a text variable
	fmt.Println("Ingesting text variable...")
	err = sys.IngestText(ctx, "Ollama is a tool for running open-source large language models locally.", "ollama-info")
	if err != nil {
		log.Fatalf("Failed to ingest text: %v", err)
	}

	// 6. Query
	queries := []string{
		"Summarize what happened to Lula Landry?",
	}

	for _, q := range queries {
		fmt.Printf("\nQuery: %s\n", q)
		response, err := sys.Query(ctx, q)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}
		fmt.Printf("Response: %s\n", response)
	}
}
