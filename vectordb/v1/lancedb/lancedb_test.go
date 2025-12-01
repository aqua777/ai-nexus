package lancedb

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aqua777/ai-nexus/vectordb/v1/schema"
	"github.com/stretchr/testify/suite"
)

type LanceDBStoreTestSuite struct {
	suite.Suite
	dbPath string
	store  *LanceDBStore
}

func TestLanceDBStoreTestSuite(t *testing.T) {
	suite.Run(t, new(LanceDBStoreTestSuite))
}

func (s *LanceDBStoreTestSuite) SetupSuite() {
	s.dbPath = s.T().TempDir()
}

func (s *LanceDBStoreTestSuite) SetupTest() {
	// Create a new store for each test to ensure isolation if needed,
	// but reusing dbPath is fine as we can use different tables or clean up.
	// Let's use a unique table name per test.
	tableName := "test_vectors_" + filepath.Base(s.T().Name())

	var err error
	s.store, err = NewLanceDBStore(s.dbPath, tableName)
	s.NoError(err)
}

func (s *LanceDBStoreTestSuite) TearDownTest() {
	if s.store != nil {
		s.store.Close()
	}
}

func (s *LanceDBStoreTestSuite) TestAddAndQuery() {
	ctx := context.Background()

	// 1. Add nodes
	nodes := []schema.Node{
		{
			ID:   "1",
			Text: "Hello world",
			Type: schema.ObjectTypeText,
			Metadata: map[string]interface{}{
				"author": "John",
				"year":   2023,
			},
			Embedding: []float32{0.1, 0.1, 0.1},
		},
		{
			ID:   "2",
			Text: "Hello space",
			Type: schema.ObjectTypeText,
			Metadata: map[string]interface{}{
				"author": "Jane",
				"year":   2024,
			},
			Embedding: []float32{0.1, 0.1, 0.2},
		},
	}

	ids, err := s.store.Add(ctx, nodes)
	s.NoError(err)
	s.Len(ids, 2)

	// 2. Query
	query := schema.VectorStoreQuery{
		Embedding: []float32{0.1, 0.1, 0.1},
		TopK:      1,
	}

	results, err := s.store.Query(ctx, query)
	s.NoError(err)
	s.Len(results, 1)
	s.Equal("1", results[0].Node.ID)
	s.Equal("Hello world", results[0].Node.Text)

	// Verify metadata reconstruction
	// Note: JSON unmarshal turns numbers into float64 usually
	s.Equal("John", results[0].Node.Metadata["author"])
}

func (s *LanceDBStoreTestSuite) TestQueryFiltering() {
	ctx := context.Background()

	nodes := []schema.Node{
		{
			ID:        "A",
			Text:      "Alpha",
			Type:      schema.ObjectTypeText,
			Metadata:  map[string]interface{}{"category": "one"},
			Embedding: []float32{1.0, 0.0},
		},
		{
			ID:        "B",
			Text:      "Beta",
			Type:      schema.ObjectTypeText,
			Metadata:  map[string]interface{}{"category": "two"},
			Embedding: []float32{0.0, 1.0},
		},
	}

	_, err := s.store.Add(ctx, nodes)
	s.NoError(err)

	// Filter by metadata
	filters := &schema.MetadataFilters{
		Filters: []schema.MetadataFilter{
			{
				Key:      "category",
				Value:    "two",
				Operator: schema.FilterOperatorEq,
			},
		},
	}

	query := schema.VectorStoreQuery{
		Embedding: []float32{0.0, 1.0}, // Match Beta
		TopK:      5,
		Filters:   filters,
	}

	results, err := s.store.Query(ctx, query)
	s.NoError(err)
	s.Len(results, 1)
	s.Equal("B", results[0].Node.ID)
}

func (s *LanceDBStoreTestSuite) TestPersistence() {
	ctx := context.Background()
	tableName := "persistence_test"

	// 1. Create and add
	store1, err := NewLanceDBStore(s.dbPath, tableName)
	s.NoError(err)

	nodes := []schema.Node{
		{ID: "p1", Text: "Persist me", Embedding: []float32{1.0, 2.0, 3.0}},
	}
	_, err = store1.Add(ctx, nodes)
	s.NoError(err)
	store1.Close()

	// 2. Reopen
	store2, err := NewLanceDBStore(s.dbPath, tableName)
	s.NoError(err)
	defer store2.Close()

	// 3. Query
	query := schema.VectorStoreQuery{
		Embedding: []float32{1.0, 2.0, 3.0},
		TopK:      1,
	}
	results, err := store2.Query(ctx, query)
	s.NoError(err)
	s.Len(results, 1)
	s.Equal("p1", results[0].Node.ID)
}
