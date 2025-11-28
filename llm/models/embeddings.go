package models

type EmbeddingsRequest struct {
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
	Content    string `json:"content"`
}

type EmbeddingsResponse struct {
	Embeddings []float32 `json:"embedding"`
}
