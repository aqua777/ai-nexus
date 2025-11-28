package chromem_test

import (
	"context"
	"testing"

	mock_llm "github.com/aqua777/ai-flow/mocks/llm"
	"github.com/aqua777/ai-flow/vectordb/v0/go-chromem"
	"github.com/aqua777/ai-flow/vectordb/v0/models"
	"github.com/stretchr/testify/suite"
)

type ChromaDBTestSuite struct {
	suite.Suite
	ctx     context.Context
	mockLLM *mock_llm.MockLLM
	db      *chromem.ChromaDB
}

func (s *ChromaDBTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockLLM = &mock_llm.MockLLM{}
	s.db = chromem.NewInMemoryDB(s.mockLLM, "test-model")
}

func (s *ChromaDBTestSuite) TestChromaDBFlow() {
	collectionName := "test_collection"
	err := s.db.CreateCollection(s.ctx, collectionName)
	s.Require().NoError(err)

	docs := []*models.Document{
		{
			ID:      "1",
			Content: "Hello world",
			Metadata: map[string]interface{}{
				"author": "Alice",
			},
			Vector: []float32{1.0, 0.0, 0.0},
		},
		{
			ID:      "2",
			Content: "Goodbye world",
			Metadata: map[string]interface{}{
				"author": "Bob",
			},
			Vector: []float32{0.0, 1.0, 0.0},
		},
	}

	err = s.db.Upsert(s.ctx, collectionName, docs)
	s.Require().NoError(err)

	// Search
	// mockEmbeddingFunc returns [1,0,0] for query "Hello".
	// Doc 1 is [1,0,0]. Dot product is 1.
	// Doc 2 is [0,1,0]. Dot product is 0.
	results, err := s.db.Search(s.ctx, collectionName, "Hello", 1)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Equal("1", results[0].Document.ID)
	s.InDelta(1.0, results[0].Score, 0.001)

	// Delete
	err = s.db.Delete(s.ctx, collectionName, []string{"1"})
	s.Require().NoError(err)

	results, err = s.db.Search(s.ctx, collectionName, "Hello", 5)
	s.Require().NoError(err)
	s.Require().Len(results, 1)
	s.Equal("2", results[0].Document.ID)

	// Delete Collection
	err = s.db.DeleteCollection(s.ctx, collectionName)
	s.Require().NoError(err)
}

func TestChromaDBTestSuite(t *testing.T) {
	suite.Run(t, new(ChromaDBTestSuite))
}
