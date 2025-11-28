package models

type Role string

const (
	UserRole Role = "user"
	AssistantRole Role = "assistant"
	SystemRole Role = "system"
)
