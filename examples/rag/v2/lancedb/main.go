package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/aqua777/ai-nexus/dotenv"
	"github.com/aqua777/ai-nexus/llm/ollama"
	"github.com/aqua777/ai-nexus/rag/v2"
	"github.com/aqua777/ai-nexus/vectordb/v1/lancedb"
)

func main() {
	ctx := context.Background()

	// 1. Initialize Clients
	// We'll use Ollama for both embedding and generation in this demo
	// Make sure Ollama is running locally (ollama serve)
	ollamaClient, err := ollama.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// We use the OpenAI client wrapper for compatibility if needed, but here we pass Ollama directly
	// as it implements the necessary interfaces.

	// 2. Configure RAG System
	config := &rag.RAGConfig{
		ChunkSize:      512,
		ChunkOverlap:   50,
		PersistPath:    "./.lancedb-data",
		CollectionName: "demo_table",
		EmbeddingModel: "mxbai-embed-large", // Make sure you have this model pulled in Ollama
		LLMModel:       "jan-v1:q6_k",            // Make sure you have this model pulled in Ollama
		FileExtensions: []string{".txt", ".md", ".not-txt"},
	}

	// 3. Initialize Vector Store (LanceDB)
	// We remove the persist path before starting to ensure a clean state for the demo
	// In a real app, you wouldn't do this.
	os.RemoveAll(config.PersistPath)

	vectorStore, err := lancedb.NewLanceDBStore(config.PersistPath, config.CollectionName)
	if err != nil {
		log.Fatalf("failed to create lancedb vector store: %w", err)
	}
	// Ensure we close the store when done
	defer vectorStore.Close()

	// 4. Initialize RAG System
	sys, err := rag.NewRAGSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	// Inject dependencies
	// We use Ollama for both embeddings and LLM
	sys.WithEmbedding(ollamaClient).WithLLM(ollamaClient).WithVectorStore(vectorStore)

	// Define Callbacks for visibility
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
	// Ingest the sample directory we created
	fmt.Println("Ingesting directory ./data ...")
	err = sys.IngestDirectory(ctx, "./data")
	if err != nil {
		log.Fatalf("Failed to ingest directory: %v", err)
	}

	// Ingest explicit text
	fmt.Println("Ingesting dynamic text...")
	err = sys.IngestText(ctx, "LanceDB is highly performant and scalable.", "lancedb-extra-info")
	if err != nil {
		log.Fatalf("Failed to ingest text: %v", err)
	}

	// 6. Query
	query := "What is LanceDB and what does it provide?"
	fmt.Printf("\nQuery: %s\n", query)

	response, err := sys.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Response: %s\n", response)
}

