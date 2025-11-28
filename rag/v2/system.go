package rag

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aqua777/ai-flow/llm/iface"
	"github.com/aqua777/ai-flow/llm/models"

	// llm_openai "github.com/aqua777/ai-flow/llm/openai"
	"github.com/aqua777/ai-flow/rag/v2/reader"
	"github.com/aqua777/ai-flow/textsplitter"
	store "github.com/aqua777/ai-flow/vectordb/v1"
	"github.com/aqua777/ai-flow/vectordb/v1/schema"

	// "github.com/aqua777/ai-flow/vectordb/v1/chromem"
	"github.com/google/uuid"
	// openai "github.com/sashabaranov/go-openai"
)

// IngestProgress reports the progress of the ingestion process.
type IngestProgress struct {
	TotalDocuments       int
	CurrentDocumentIndex int
	TotalChunksInDoc     int
	CurrentChunkIndex    int
	Message              string
}

// IngestionCallbacks defines customizable event handlers for the ingestion process.
type IngestionCallbacks struct {
	OnIngestStarted   func(totalDocs int)
	OnIngestProgress  func(progress IngestProgress)
	OnIngestCompleted func()
	OnIngestError     func(err error)
}

// RAGConfig holds configuration for the RAG system.
type RAGConfig struct {
	OpenAIKey      string
	OpenAIBaseURL  string // Optional: for using other OpenAI-compatible APIs
	LLMModel       string
	EmbeddingModel string
	ChunkSize      int
	ChunkOverlap   int
	TopK           int
	PersistPath    string   // Path to persist vector store. Empty for in-memory.
	CollectionName string   // Name of the vector store collection.
	FileExtensions []string // File extensions to process (e.g., ".txt", ".md")
}

// RAGSystem encapsulates the RAG pipeline components.
type RAGSystem struct {
	Config      *RAGConfig
	Embedder    iface.LLM
	LLM         iface.LLM
	VectorStore store.VectorStore
	QueryEngine *RetrieverQueryEngine
	Splitter    *textsplitter.SentenceSplitter
	Callbacks   IngestionCallbacks
}

// NewRAGSystem creates a new RAGSystem with the provided configuration.
func NewRAGSystem(config *RAGConfig) (*RAGSystem, error) {
	// Set Defaults
	if config.LLMModel == "" {
		config.LLMModel = "gpt-3.5-turbo"
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = "text-embedding-3-small"
	}
	if config.ChunkSize <= 0 {
		config.ChunkSize = 1024
	}
	if config.ChunkOverlap <= 0 {
		config.ChunkOverlap = 200
	}
	if config.TopK <= 0 {
		config.TopK = 3
	}
	if config.CollectionName == "" {
		config.CollectionName = "documents"
	}
	if len(config.FileExtensions) == 0 {
		config.FileExtensions = []string{".txt", ".md"}
	}

	// // Vector Store
	// // ChromemStore implements VectorStore interface
	// vectorStore, err := chromem.NewChromemStore(config.PersistPath, config.CollectionName)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create vector store: %w", err)
	// }

	// Splitter
	splitter := textsplitter.NewSentenceSplitter(config.ChunkSize, config.ChunkOverlap, nil, nil)

	sys := &RAGSystem{
		Config: config,
		// VectorStore: vectorStore,
		Splitter: splitter,
	}

	return sys, nil
}

// WithEmbedding allows injecting a custom embedding model.
// This must be called before usage (Ingest/Query) to ensure the pipeline is correctly set up.
func (s *RAGSystem) WithEmbedding(embedder iface.LLM) *RAGSystem {
	s.Embedder = embedder
	return s
}

// WithLLM allows injecting a custom LLM.
// This must be called before usage (Query) to ensure the pipeline is correctly set up.
func (s *RAGSystem) WithLLM(llmModel iface.LLM) *RAGSystem {
	s.LLM = llmModel
	return s
}

func (s *RAGSystem) WithVectorStore(vectorStore store.VectorStore) *RAGSystem {
	s.VectorStore = vectorStore
	return s
}

func (s *RAGSystem) WithOnIngestStarted(callback func(totalDocs int)) *RAGSystem {
	s.Callbacks.OnIngestStarted = callback
	return s
}

func (s *RAGSystem) WithOnIngestProgress(callback func(progress IngestProgress)) *RAGSystem {
	s.Callbacks.OnIngestProgress = callback
	return s
}

func (s *RAGSystem) WithOnIngestCompleted(callback func()) *RAGSystem {
	s.Callbacks.OnIngestCompleted = callback
	return s
}

func (s *RAGSystem) WithOnIngestError(callback func(err error)) *RAGSystem {
	s.Callbacks.OnIngestError = callback
	return s
}

// bootstrap ensures that the QueryEngine and other dependent components are initialized.
// It should be called lazily or explicitly before operations that need them.
func (s *RAGSystem) bootstrap() error {
	if s.Embedder == nil {
		return fmt.Errorf("embedding model is not initialized, use WithEmbedding()")
	}
	if s.LLM == nil {
		return fmt.Errorf("LLM is not initialized, use WithLLM()")
	}
	if s.VectorStore == nil {
		return fmt.Errorf("vector store is not initialized, use WithVectorStore()")
	}

	// Re-initialize QueryEngine if it doesn't exist or if components changed
	// For simplicity, we just recreate it if it's nil or if we want to be safe.
	// Given the chainable nature, we can't easily know when "configuration" is done.
	// So we'll check if QueryEngine is nil or if we want to force update.
	// Let's just create it if it's nil.
	if s.QueryEngine == nil {
		retriever := NewVectorRetriever(s.VectorStore, s.Embedder, s.Config.EmbeddingModel, s.Config.TopK)
		synthesizer := NewSimpleSynthesizer(s.LLM, s.Config.LLMModel)
		s.QueryEngine = NewRetrieverQueryEngine(retriever, synthesizer)
	}
	return nil
}

// IngestDirectory loads documents from inputDir, chunks them, embeds them, and stores them in the vector store.
func (s *RAGSystem) IngestDirectory(ctx context.Context, inputDir string) error {
	if err := s.bootstrap(); err != nil {
		return err
	}

	// 1. Load Data
	// We unpack FileExtensions to pass as variadic arguments
	// reader.NewSimpleDirectoryReader expects specific extensions
	// Actually NewSimpleDirectoryReader takes (dir string, ext ...string)
	docReader := reader.NewSimpleDirectoryReader(inputDir, s.Config.FileExtensions...)

	// The SimpleDirectoryReader currently returns []schema.Node (which act as documents).
	// We need to convert them to []schema.Document.
	nodes, err := docReader.LoadData()
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	if len(nodes) == 0 {
		log.Printf("No documents matching extensions %v found in %s", s.Config.FileExtensions, inputDir)
		return nil
	}

	// Convert nodes to docs
	var docs []schema.Document
	for _, node := range nodes {
		docs = append(docs, schema.Document{
			ID:       node.ID,
			Text:     node.Text,
			Metadata: node.Metadata,
		})
	}

	return s.ingestDocuments(ctx, docs)
}

// IngestText accepts a raw string of text, creates a document from it, and ingests it.
func (s *RAGSystem) IngestText(ctx context.Context, text string, sourceID string) error {
	if err := s.bootstrap(); err != nil {
		return err
	}

	if sourceID == "" {
		sourceID = uuid.New().String()
	}
	doc := schema.Document{
		ID:   sourceID,
		Text: text,
		Metadata: map[string]interface{}{
			"source_id": sourceID,
			"source":    "text_variable",
		},
	}
	return s.ingestDocuments(ctx, []schema.Document{doc})
}

// IngestFile reads a single file and ingests it.
func (s *RAGSystem) IngestFile(ctx context.Context, filePath string) error {
	if err := s.bootstrap(); err != nil {
		return err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	doc := schema.Document{
		ID:   filePath,
		Text: string(content),
		Metadata: map[string]interface{}{
			"source_id": filePath,
			"filename":  filePath,
		},
	}
	return s.ingestDocuments(ctx, []schema.Document{doc})
}

// ingestDocuments handles the common logic of splitting, embedding, and adding documents to the store.
func (s *RAGSystem) ingestDocuments(ctx context.Context, docs []schema.Document) error {
	totalDocs := len(docs)
	if s.Callbacks.OnIngestStarted != nil {
		s.Callbacks.OnIngestStarted(totalDocs)
	}

	// 2. Split and Embed
	var allNodes []schema.Node
	for docIdx, doc := range docs {
		chunks := s.Splitter.SplitText(doc.Text)
		totalChunks := len(chunks)

		for i, chunk := range chunks {
			if s.Callbacks.OnIngestProgress != nil {
				s.Callbacks.OnIngestProgress(IngestProgress{
					TotalDocuments:       totalDocs,
					CurrentDocumentIndex: docIdx,
					TotalChunksInDoc:     totalChunks,
					CurrentChunkIndex:    i,
					Message:              fmt.Sprintf("Processing document %s, chunk %d/%d", doc.ID, i+1, totalChunks),
				})
			}

			// Create node
			node := schema.Node{
				ID:   fmt.Sprintf("%s-chunk-%d", doc.ID, i),
				Text: chunk,
				Type: schema.ObjectTypeText,
				Metadata: map[string]interface{}{
					"source_id": doc.ID,
				},
			}
			// Copy over document metadata if it exists
			if doc.Metadata != nil {
				for k, v := range doc.Metadata {
					node.Metadata[k] = v
				}
			}

			// Generate embedding explicitly
			resp, err := s.Embedder.Embeddings(ctx, &models.EmbeddingsRequest{
				Content: chunk,
				Model:   s.Config.EmbeddingModel,
			})
			if err != nil {
				err = fmt.Errorf("failed to get embedding for chunk %d of doc %s: %w", i, doc.ID, err)
				if s.Callbacks.OnIngestError != nil {
					s.Callbacks.OnIngestError(err)
				}
				return err
			}

			// Convert float32 to float64
			embedding := make([]float64, len(resp.Embeddings))
			for j, v := range resp.Embeddings {
				embedding[j] = float64(v)
			}

			node.Embedding = embedding
			allNodes = append(allNodes, node)
		}
	}

	// 3. Ingest
	if len(allNodes) > 0 {
		_, err := s.VectorStore.Add(ctx, allNodes)
		if err != nil {
			err = fmt.Errorf("failed to add nodes to vector store: %w", err)
			if s.Callbacks.OnIngestError != nil {
				s.Callbacks.OnIngestError(err)
			}
			return err
		}
	}

	if s.Callbacks.OnIngestCompleted != nil {
		s.Callbacks.OnIngestCompleted()
	}

	return nil
}

// Query executes a query against the RAG system and returns the response.
func (s *RAGSystem) Query(ctx context.Context, queryStr string) (string, error) {
	if err := s.bootstrap(); err != nil {
		return "", err
	}

	response, err := s.QueryEngine.Query(ctx, schema.QueryBundle{QueryString: queryStr})
	if err != nil {
		return "", err
	}
	return response.Response, nil
}
