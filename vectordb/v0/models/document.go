package models

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Document struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title,omitempty"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Vector   []float32              `json:"vector,omitempty"`
}

type SearchResult struct {
	Document *Document `json:"document"`
	Score    float32   `json:"score"`
}

// DocumentFromFile creates a Document from a file.
// It reads the file content and sets it as the Document Content.
// The title will be set to the base file name.
func DocumentFromFile(path string) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fileName := filepath.Base(path)

	return &Document{
		ID:       uuid.New().String(),
		Title:    fileName,
		Content:  string(content),
		Metadata: map[string]interface{}{},
	}, nil
}
