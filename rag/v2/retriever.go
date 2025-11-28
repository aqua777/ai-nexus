package rag

import (
	"context"
	"fmt"

	"github.com/aqua777/ai-flow/llm/iface"
	"github.com/aqua777/ai-flow/llm/models"
	"github.com/aqua777/ai-flow/vectordb/v1/schema"
	"github.com/aqua777/ai-flow/vectordb/v1"
)

// VectorRetriever retrieves relevant nodes using a vector store and embedding model.
type VectorRetriever struct {
	vectorStore        store.VectorStore
	embedder           iface.LLM
	embeddingModelName string
	topK               int
}

// NewVectorRetriever creates a new VectorRetriever.
func NewVectorRetriever(vectorStore store.VectorStore, embedder iface.LLM, embeddingModelName string, topK int) *VectorRetriever {
	return &VectorRetriever{
		vectorStore:        vectorStore,
		embedder:           embedder,
		embeddingModelName: embeddingModelName,
		topK:               topK,
	}
}

func (r *VectorRetriever) Retrieve(ctx context.Context, query schema.QueryBundle) ([]schema.NodeWithScore, error) {
	resp, err := r.embedder.Embeddings(ctx, &models.EmbeddingsRequest{
		Content: query.QueryString,
		Model:   r.embeddingModelName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Convert float32 to float64
	queryEmbedding := make([]float64, len(resp.Embeddings))
	for i, v := range resp.Embeddings {
		queryEmbedding[i] = float64(v)
	}

	storeQuery := schema.VectorStoreQuery{
		Embedding: queryEmbedding,
		TopK:      r.topK,
		Filters:   query.Filters,
	}

	nodes, err := r.vectorStore.Query(ctx, storeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query vector store: %w", err)
	}

	return nodes, nil
}
