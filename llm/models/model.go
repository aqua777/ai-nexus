package models

type Model struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Model string `json:"model"`
	Description string `json:"description"`
	ContextSize int `json:"context_size"`
}
