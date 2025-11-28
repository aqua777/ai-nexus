package openai

type ModelResponse struct {
	ID	string `json:"id"`
	Object string `json:"object"`
	Created int64 `json:"created"`
	OwnedBy string `json:"owned_by"`
}
