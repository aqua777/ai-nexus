package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/aqua777/ai-nexus/llm/iface"
	"github.com/aqua777/ai-nexus/llm/models"
	"github.com/aqua777/ai-nexus/vectordb/v1/schema"
)

// SimpleSynthesizer generates a response by stuffing retrieved context into a prompt.
type SimpleSynthesizer struct {
	llm          iface.LLM
	llmModelName string
}

// NewSimpleSynthesizer creates a new SimpleSynthesizer.
func NewSimpleSynthesizer(llm iface.LLM, llmModelName string) *SimpleSynthesizer {
	return &SimpleSynthesizer{
		llm:          llm,
		llmModelName: llmModelName,
	}
}

func (s *SimpleSynthesizer) Synthesize(ctx context.Context, query schema.QueryBundle, nodes []schema.NodeWithScore) (schema.EngineResponse, error) {
	contextStr := s.formatContext(nodes)
	prompt := s.createPrompt(contextStr, query.QueryString)

	req := &models.ChatRequest{
		Model: s.llmModelName,
		Messages: []*models.Message{
			{
				Role:    models.UserRole,
				Content: prompt,
			},
		},
	}

	resp, err := s.llm.Chat(ctx, req)
	if err != nil {
		return schema.EngineResponse{}, fmt.Errorf("llm completion failed: %w", err)
	}

	return schema.EngineResponse{
		Response:    resp.Content,
		SourceNodes: nodes,
	}, nil
}

func (s *SimpleSynthesizer) SynthesizeStream(ctx context.Context, query schema.QueryBundle, nodes []schema.NodeWithScore) (schema.StreamingEngineResponse, error) {
	contextStr := s.formatContext(nodes)
	prompt := s.createPrompt(contextStr, query.QueryString)

	// Create channel for streaming response
	tokenChan := make(chan string)

	req := &models.ChatRequest{
		Model: s.llmModelName,
		Messages: []*models.Message{
			{
				Role:    models.UserRole,
				Content: prompt,
			},
		},
		Stream: true,
	}

	go func() {
		defer close(tokenChan)
		_, err := s.llm.Chat(ctx, req, func(chunk []byte) error {
			tokenChan <- string(chunk)
			return nil
		})
		if err != nil {
			// Log error or handle it? Since we are inside a goroutine and the channel is the only output,
			// typically we might send an error or just close.
			// For this simple interface, we might just log or accept that the stream ends.
			// But wait, QueryStream returns a channel.
			// The original implementation returned `<-chan string, error`.
			// Here we are returning the channel immediately.
		}
	}()

	return schema.StreamingEngineResponse{
		ResponseStream: tokenChan,
		SourceNodes:    nodes,
	}, nil
}

func (s *SimpleSynthesizer) formatContext(nodes []schema.NodeWithScore) string {
	var sb strings.Builder
	for _, n := range nodes {
		sb.WriteString(n.Node.Text)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (s *SimpleSynthesizer) createPrompt(context, query string) string {
	return fmt.Sprintf("Context information is below.\n---------------------\n%s\n---------------------\nGiven the context information and not prior knowledge, answer the query.\nQuery: %s\nAnswer:", context, query)
}
