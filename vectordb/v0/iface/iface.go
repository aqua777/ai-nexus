package iface

import (
	"context"

	"github.com/aqua777/ai-nexus/vectordb/v0/models"
)

type VectorDB interface {
	// CreateCollection creates a new collection with the given name.
	CreateCollection(ctx context.Context, name string) error

	// DeleteCollection deletes an existing collection by name.
	DeleteCollection(ctx context.Context, name string) error

	// Upsert adds or updates documents in the specified collection.
	Upsert(ctx context.Context, collectionName string, documents []*models.Document) error

	// Search performs a semantic search using a query string and returns the top k results.
	Search(ctx context.Context, collectionName string, query string, k int) ([]*models.SearchResult, error)

	// Delete removes documents from the collection by their IDs.
	Delete(ctx context.Context, collectionName string, documentIDs []string) error
}
