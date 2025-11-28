package ollama

import (
	"time"
)

type OllamaModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParamaterSize     int64    `json:"paramater_size"`
	QuantizationLevel string   `json:"quantization"`
}

type OllamaModel struct {
	Name         string             `json:"name"`
	Model        string             `json:"model"`
	ModelFile    string             `json:"modelfile"`
	Parameters   string             `json:"parameters"`
	Template     string             `json:"template"`
	Size         int64              `json:"size"`
	Digest       string             `json:"digest"`
	Details      OllamaModelDetails `json:"details"`
	ModelInfo    map[string]any     `json:"model_info"`
	Capabilities []string           `json:"capabilities"`
	ModifiedAt   time.Time          `json:"modified_at"`
}

type ListModelResponse struct {
	Models []*OllamaModel `json:"models"`
}

