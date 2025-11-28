package models

type RequestOptions struct {
	Temperature float64 `json:"temperature"`
	TopP float64 `json:"top_p"`
	MaxTokens int `json:"max_tokens"`
	TopK int `json:"top_k"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty float64 `json:"presence_penalty"`
}

func (o *RequestOptions) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	if o == nil {
		return result
	}
	if o.Temperature != 0 {
		result["temperature"] = o.Temperature
	}
	if o.TopP != 0 {
		result["top_p"] = o.TopP
	}
	if o.MaxTokens != 0 {
		result["max_tokens"] = o.MaxTokens
	}
	if o.TopK != 0 {
		result["top_k"] = o.TopK
	}
	if o.FrequencyPenalty != 0 {
		result["frequency_penalty"] = o.FrequencyPenalty
	}
	if o.PresencePenalty != 0 {
		result["presence_penalty"] = o.PresencePenalty
	}
	return result
}
