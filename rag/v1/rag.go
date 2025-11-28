package rag

import (
	"context"
	"fmt"

	"github.com/aqua777/ai-nexus/textsplitter"
	"github.com/aqua777/ai-nexus/vectordb/v0/iface"
	"github.com/aqua777/ai-nexus/vectordb/v0/models"
)

// Service provides RAG (Retrieval Augmented Generation) capabilities
// by coordinating document chunking, embedding, and storage/retrieval.
type Service struct {
	vectorDB iface.VectorDB
	splitter textsplitter.TextSplitter
}

// NewService creates a new RAG service.
func NewService(vdb iface.VectorDB, splitter textsplitter.TextSplitter) *Service {
	return &Service{
		vectorDB: vdb,
		splitter: splitter,
	}
}

// CreateCollection creates a new collection in the vector database.
func (s *Service) CreateCollection(ctx context.Context, name string) error {
	return s.vectorDB.CreateCollection(ctx, name)
}

// DeleteCollection deletes a collection from the vector database.
func (s *Service) DeleteCollection(ctx context.Context, name string) error {
	return s.vectorDB.DeleteCollection(ctx, name)
}

// Ingest processes a single document: chunks it and stores it in the vector database.
func (s *Service) Ingest(ctx context.Context, collectionName string, doc *models.Document) error {
	// 1. Chunk the document
	textChunks := s.splitter.SplitText(doc.Content)
	var chunks []*models.Document

	for i, textChunk := range textChunks {
		// Create a new document for the chunk
		chunkID := fmt.Sprintf("%s_chunk_%d", doc.ID, i)

		// Copy metadata and add chunk-specific metadata
		metadata := make(map[string]interface{})
		for k, v := range doc.Metadata {
			metadata[k] = v
		}
		metadata["source_id"] = doc.ID
		metadata["chunk_index"] = i

		chunk := &models.Document{
			ID:       chunkID,
			Content:  textChunk,
			Metadata: metadata,
		}
		chunks = append(chunks, chunk)
	}

	if len(chunks) == 0 {
		return nil
	}

	// 2. Store chunks in VectorDB
	// The VectorDB implementation is expected to handle embedding if vectors are missing.
	if err := s.vectorDB.Upsert(ctx, collectionName, chunks); err != nil {
		return fmt.Errorf("failed to upsert chunks: %w", err)
	}

	return nil
}

// BatchIngest processes multiple documents: chunks them and stores them in the vector database.
func (s *Service) BatchIngest(ctx context.Context, collectionName string, docs []*models.Document) error {
	var allChunks []*models.Document

	// 1. Chunk each document
	for _, doc := range docs {
		textChunks := s.splitter.SplitText(doc.Content)
		for i, textChunk := range textChunks {
			chunkID := fmt.Sprintf("%s_chunk_%d", doc.ID, i)

			metadata := make(map[string]interface{})
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
			metadata["source_id"] = doc.ID
			metadata["chunk_index"] = i

			chunk := &models.Document{
				ID:       chunkID,
				Content:  textChunk,
				Metadata: metadata,
			}
			allChunks = append(allChunks, chunk)
		}
	}

	if len(allChunks) == 0 {
		return nil
	}

	// 2. Store chunks in VectorDB
	if err := s.vectorDB.Upsert(ctx, collectionName, allChunks); err != nil {
		return fmt.Errorf("failed to upsert chunks: %w", err)
	}

	return nil
}

// Retrieve searches for relevant documents in the specified collection using the query.
func (s *Service) Retrieve(ctx context.Context, collectionName string, query string, k int) ([]*models.SearchResult, error) {
	return s.vectorDB.Search(ctx, collectionName, query, k)
}
