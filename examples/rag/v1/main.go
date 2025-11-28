package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aqua777/ai-nexus/llm/ollama"
	"github.com/aqua777/ai-nexus/rag/v1"
	"github.com/aqua777/ai-nexus/textsplitter"
	"github.com/aqua777/ai-nexus/vectordb/v0/go-chromem"
	vdb_models "github.com/aqua777/ai-nexus/vectordb/v0/models"

	ssExample "github.com/aqua777/ai-nexus/examples/textsplitter/sentence-splitter/funcs"
)

var _ = os.Args

func testSplitters(text string) {
	ssExample.TestSplitters(text)
	os.Exit(0)
}

func main() {
	fileName := flag.String("file", "", "The file to ingest")
	flag.Parse()
	if *fileName == "" {
		log.Fatalf("Please provide a file name using -file flag")
	}
	// 1. Ingest Document
	doc, err := vdb_models.DocumentFromFile(*fileName)
	if err != nil {
		log.Fatalf("Failed to load document from file: %v", err)
	}

	// testSplitters(doc.Content)

	// 2. Initialize TextSplitter (using SentenceSplitter)
	// Use TikToken tokenizer for better token counting (compatible with OpenAI models and similar)
	// tokenizer := textsplitter.NewSimpleTokenizer()
	tokenizer, err := textsplitter.NewTikTokenTokenizer("gpt-3.5-turbo")
	if err != nil {
		log.Fatalf("Failed to create tokenizer %p, %v", tokenizer, err)
	}

	strategy, err := textsplitter.NewNeurosnapSplitterStrategy(nil)
	if err != nil {
		log.Fatalf("Failed to create neurosnap strategy %p, %v", strategy, err)
	}

	splitter := textsplitter.NewSentenceSplitter(200, 20, nil, strategy)

	ctx := context.Background()
	llmClient, err := ollama.NewClient()
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	// 3. Initialize VectorDB (InMemory Chroma)
	// We use "nomic-embed-text" as the embedding model. Ensure you have it pulled in Ollama: `ollama pull nomic-embed-text`
	embeddingModel := "nomic-embed-text"
	vdb := chromem.NewInMemoryDB(llmClient, embeddingModel)

	// 4. Initialize RAG Service
	ragService := rag.NewService(vdb, splitter)

	// 5. Create Collection
	collectionName := "demo_collection"
	if err := ragService.CreateCollection(ctx, collectionName); err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}

	fmt.Println("Ingesting document...")
	if err := ragService.Ingest(ctx, collectionName, doc); err != nil {
		log.Fatalf("Failed to ingest document: %v", err)
	}
	fmt.Println("Document ingested successfully.")

	// 7. Retrieve
	query := "Describe Robin's appearance."
	fmt.Printf("Querying: %s\n", query)
	results, err := ragService.Retrieve(ctx, collectionName, query, 3)
	if err != nil {
		log.Fatalf("Failed to retrieve: %v", err)
	}

	fmt.Printf("Found %d results:\n", len(results))
	for i, res := range results {
		fmt.Printf("[%d] Score: %.4f\nContent(%d chars): %s\n\n", i+1, res.Score, len(res.Document.Content), res.Document.Content)
	}
}
