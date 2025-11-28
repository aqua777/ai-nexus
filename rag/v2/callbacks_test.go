package rag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aqua777/ai-flow/vectordb/v1"
)

type CallbacksTestSuite struct {
	suite.Suite
}

func TestCallbacksTestSuite(t *testing.T) {
	suite.Run(t, new(CallbacksTestSuite))
}

func (s *CallbacksTestSuite) TestIngestionCallbacks() {
	ctx := context.Background()
	
	// 1. Setup
	mockLLM := &MockLLM{
		Embedding: []float32{0.1, 0.2, 0.3},
	}
	vectorStore := store.NewSimpleVectorStore()
	
	config := &RAGConfig{
		ChunkSize:    10,
		ChunkOverlap: 0,
	}
	ragSystem, err := NewRAGSystem(config)
	s.NoError(err)
	
	ragSystem.WithEmbedding(mockLLM).WithLLM(mockLLM).WithVectorStore(vectorStore)
	
	// 2. Define Callbacks
	var started bool
	var progressCount int
	var completed bool
	var errResult error
	
	ragSystem.WithOnIngestStarted(func(totalDocs int) {
		started = true
		s.Equal(1, totalDocs)
	}).WithOnIngestProgress(func(p IngestProgress) {
		progressCount++
		s.Equal(1, p.TotalDocuments)
		s.Greater(p.TotalChunksInDoc, 0)
	}).WithOnIngestCompleted(func() {
		completed = true
	}).WithOnIngestError(func(err error) {
		errResult = err
	})
	
	// 3. Execute
	err = ragSystem.IngestText(ctx, "This is a test document that should be split into chunks.", "test-id")
	s.NoError(err)
	
	// 4. Verify
	s.True(started, "OnIngestStarted should be called")
	s.True(completed, "OnIngestCompleted should be called")
	s.Nil(errResult, "OnIngestError should not be called")
	s.Greater(progressCount, 0, "OnIngestProgress should be called at least once")
}


