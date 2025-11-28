package models

import "time"

type GenerateRequest struct {
	Model string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool `json:"stream"`
	Options RequestOptions `json:"options,omitempty"`
}

type GenerateResponse struct {
	Text string `json:"text"`
	Model string `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	PromptTokens int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens int `json:"total_tokens"`
}
