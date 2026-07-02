package llm

import "time"

// GeminiClient calls Google Gemini through its OpenAI-compatible API.
type GeminiClient struct {
	*chatCompletionClient
}

// NewGeminiClient creates a real LLM client for Gemini-compatible chat.
func NewGeminiClient(apiKey, model, baseURL string, timeout time.Duration) (*GeminiClient, error) {
	client, err := newChatCompletionClient("gemini", apiKey, model, baseURL, timeout)
	if err != nil {
		return nil, err
	}
	return &GeminiClient{chatCompletionClient: client}, nil
}
