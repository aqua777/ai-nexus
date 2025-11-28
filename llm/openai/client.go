package openai

import (
	"context"
	"errors"
	"io"

	"github.com/aqua777/ai-flow/llm/iface"
	"github.com/aqua777/ai-flow/llm/models"
	openai "github.com/sashabaranov/go-openai"
)

const (
	OpenAI_API_URL_v1 = "https://api.openai.com/v1"
)

type Client struct {
	client *openai.Client
}

// Ensure Client implements iface.LLM
var _ iface.LLM = (*Client)(nil)

func NewClient(optionalConfig ...*models.LLMConfig) (*Client, error) {
	// var config *models.LLMConfig
	// if len(optionalConfig) > 0 && optionalConfig[0] != nil {
	// 	config = optionalConfig[0]
	// } else {
	// 	config = &models.LLMConfig{}
	// }

	config := models.OptionalConfig(optionalConfig).GetConfig(models.OPENAI)
	openaiConfig := openai.DefaultConfig(config.ApiKey)
	openaiConfig.BaseURL = config.Url
	client := openai.NewClientWithConfig(openaiConfig)

	return &Client{
		client: client,
	}, nil
}

func NewClientWithOpenAIClient(client *openai.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) ListModels(ctx context.Context) ([]*models.Model, error) {
	resp, err := c.client.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	var result []*models.Model
	for _, m := range resp.Models {
		result = append(result, &models.Model{
			ID:    m.ID,
			Name:  m.ID, // OpenAI doesn't have a separate name field usually
			Model: m.ID,
		})
	}
	return result, nil
}

func (c *Client) Generate(ctx context.Context, r *models.GenerateRequest) (*models.GenerateResponse, error) {
	// OpenAI Chat Completion as Generate
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: r.Prompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    r.Model,
		Messages: messages,
	}
	// Map options if needed, skipping for now as options are generic map

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	return &models.GenerateResponse{
		Text:             resp.Choices[0].Message.Content,
		Model:            resp.Model,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}, nil
}

func (c *Client) Chat(ctx context.Context, r *models.ChatRequest, stream ...func(chunk []byte) error) (*models.ChatResponse, error) {
	openaiMessages := make([]openai.ChatCompletionMessage, len(r.Messages))
	for i, msg := range r.Messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    r.Model,
		Messages: openaiMessages,
		Stream:   len(stream) > 0 && stream[0] != nil,
	}

	if req.Stream {
		return c.streamChat(ctx, req, stream[0])
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	content := resp.Choices[0].Message.Content
	
	return &models.ChatResponse{
		Content: content,
		Metadata: &models.ChatResponseMetadata{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (c *Client) streamChat(ctx context.Context, req openai.ChatCompletionRequest, callback func(chunk []byte) error) (*models.ChatResponse, error) {
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	var fullContent string

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				fullContent += delta
				if err := callback([]byte(delta)); err != nil {
					return nil, err
				}
			}
		}
	}

	// Streaming response usually doesn't have full usage stats in the stream chunks easily aggregated 
	// without counting tokens ourselves, returning basic response.
	return &models.ChatResponse{
		Content: fullContent,
	}, nil
}

func (c *Client) Embeddings(ctx context.Context, cr *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
	model := openai.EmbeddingModel(cr.Model)
	if model == "" {
		model = openai.SmallEmbedding3
	}

	resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{cr.Content},
		Model: model,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embeddings returned")
	}

	return &models.EmbeddingsResponse{
		Embeddings: resp.Data[0].Embedding,
	}, nil
}

