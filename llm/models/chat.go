package models

type Message struct {
	Role     Role   `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
}

type ChatRequest struct {
	Model    string                        `json:"model"`
	Messages []*Message                    `json:"messages"`
	Stream   bool                          `json:"stream"`
	Options  RequestOptions `json:"options"`
}

type ChatResponseMetadata struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	Content   string                `json:"content"`
	Reasoning string                `json:"reasoning"`
	Metadata  *ChatResponseMetadata `json:"metadata"`
}
