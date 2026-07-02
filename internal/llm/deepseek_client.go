package llm

import "time"

// DeepSeekClient calls DeepSeek through its OpenAI-compatible API.
type DeepSeekClient struct {
	*chatCompletionClient
}

// NewDeepSeekClient creates a real LLM client for DeepSeek-compatible chat.
func NewDeepSeekClient(apiKey, model, baseURL string, timeout time.Duration) (*DeepSeekClient, error) {
	client, err := newChatCompletionClient("deepseek", apiKey, model, baseURL, timeout)
	if err != nil {
		return nil, err
	}
	return &DeepSeekClient{chatCompletionClient: client}, nil
}
