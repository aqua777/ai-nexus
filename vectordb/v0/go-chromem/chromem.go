package chromem

import (
	"context"
	"fmt"

	llm_iface "github.com/aqua777/ai-flow/llm/iface"
	llm_models "github.com/aqua777/ai-flow/llm/models"
	"github.com/aqua777/ai-flow/vectordb/v0/iface"
	"github.com/aqua777/ai-flow/vectordb/v0/models"
	chromem "github.com/philippgille/chromem-go"
)

type ChromaDB struct {
	db             *chromem.DB
	llmClient      llm_iface.LLM
	embeddingModel string
}

// NewInMemoryDB creates a new in-memory ChromaDB instance.
// llmClient: The LLM provider to use for embeddings.
// embeddingModel: The name of the model to use for embeddings (e.g. "text-embedding-3-small").
func NewInMemoryDB(llmClient llm_iface.LLM, embeddingModel string) *ChromaDB {
	return &ChromaDB{
		db:             chromem.NewDB(),
		llmClient:      llmClient,
		embeddingModel: embeddingModel,
	}
}

// NewPersistentDB creates a new persistent ChromaDB instance.
// llmClient: The LLM provider to use for embeddings.
// embeddingModel: The name of the model to use for embeddings (e.g. "text-embedding-3-small").
func NewPersistentDB(path string, llmClient llm_iface.LLM, embeddingModel string) (*ChromaDB, error) {
	db, err := chromem.NewPersistentDB(path, false)
	if err != nil {
		return nil, err
	}
	return &ChromaDB{
		db:             db,
		llmClient:      llmClient,
		embeddingModel: embeddingModel,
	}, nil
}

// Ensure ChromaDB implements VectorDB
var _ iface.VectorDB = (*ChromaDB)(nil)

// getEmbeddingFunc returns an adapter that converts chromem's embedding request
// to our LLM interface's Embeddings call.
func (c *ChromaDB) getEmbeddingFunc() chromem.EmbeddingFunc {
	if c.llmClient == nil {
		return nil
	}
	return func(ctx context.Context, text string) ([]float32, error) {
		resp, err := c.llmClient.Embeddings(ctx, &llm_models.EmbeddingsRequest{
			Model:   c.embeddingModel,
			Content: text,
		})
		if err != nil {
			// Add logging to show more details about the error
			fmt.Printf("Error generating embedding for text (length: %d): %v\n", len(text), err)
			return nil, err
		}
		return resp.Embeddings, nil
	}
}

func (c *ChromaDB) CreateCollection(ctx context.Context, name string) error {
	_, err := c.db.CreateCollection(name, nil, c.getEmbeddingFunc())
	return err
}

func (c *ChromaDB) DeleteCollection(ctx context.Context, name string) error {
	return c.db.DeleteCollection(name)
}

func (c *ChromaDB) Upsert(ctx context.Context, collectionName string, documents []*models.Document) error {
	col := c.db.GetCollection(collectionName, nil)
	if col == nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}

	chromaDocs := make([]chromem.Document, len(documents))
	for i, doc := range documents {
		meta := make(map[string]string)
		for k, v := range doc.Metadata {
			meta[k] = fmt.Sprintf("%v", v)
		}

		chromaDocs[i] = chromem.Document{
			ID:        doc.ID,
			Content:   doc.Content,
			Metadata:  meta,
			Embedding: doc.Vector,
		}
	}

	// Concurrency 1 for safety/simplicity, can be increased.
	return col.AddDocuments(ctx, chromaDocs, 1)
}

func (c *ChromaDB) Search(ctx context.Context, collectionName string, query string, k int) ([]*models.SearchResult, error) {
	col := c.db.GetCollection(collectionName, nil)
	if col == nil {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	count := col.Count()
	if count == 0 {
		return []*models.SearchResult{}, nil
	}
	if k > count {
		k = count
	}

	// chromem-go Query embeds the query string using the collection's embedding function.
	results, err := col.Query(ctx, query, k, nil, nil)
	if err != nil {
		return nil, err
	}

	searchResults := make([]*models.SearchResult, len(results))
	for i, res := range results {
		meta := make(map[string]interface{})
		for k, v := range res.Metadata {
			meta[k] = v
		}

		searchResults[i] = &models.SearchResult{
			Document: &models.Document{
				ID:       res.ID,
				Content:  res.Content,
				Metadata: meta,
				Vector:   res.Embedding,
			},
			Score: res.Similarity,
		}
	}
	return searchResults, nil
}

func (c *ChromaDB) Delete(ctx context.Context, collectionName string, documentIDs []string) error {
	col := c.db.GetCollection(collectionName, nil)
	if col == nil {
		return fmt.Errorf("collection %s not found", collectionName)
	}
	return col.Delete(ctx, nil, nil, documentIDs...)
}
