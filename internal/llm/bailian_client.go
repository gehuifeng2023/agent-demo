package llm

import "time"

type bailianMessage = chatCompletionMessage
type bailianRequest = chatCompletionRequest
type bailianResponse = chatCompletionResponse

// BailianClient calls Alibaba Bailian through the DashScope OpenAI-compatible API.
type BailianClient struct {
	*chatCompletionClient
}

// NewBailianClient creates a real LLM client for DashScope-compatible chat.
func NewBailianClient(apiKey, model, baseURL string, timeout time.Duration) (*BailianClient, error) {
	client, err := newChatCompletionClient("bailian", apiKey, model, baseURL, timeout)
	if err != nil {
		return nil, err
	}
	return &BailianClient{chatCompletionClient: client}, nil
}
