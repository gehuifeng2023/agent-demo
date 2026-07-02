package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"agent-demo/internal/prompt"
)

type chatCompletionClient struct {
	provider   string
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []chatCompletionMessage `json:"messages"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatCompletionMessage `json:"message"`
	} `json:"choices"`
}

func newChatCompletionClient(provider, apiKey, model, baseURL string, timeout time.Duration) (*chatCompletionClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("%s api key is required", provider)
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("%s model is required", provider)
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("%s base url is required", provider)
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &chatCompletionClient{
		provider: strings.TrimSpace(provider),
		apiKey:   strings.TrimSpace(apiKey),
		model:    strings.TrimSpace(model),
		baseURL:  strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Generate sends a single-turn chat request and returns the provider answer.
func (c *chatCompletionClient) Generate(ctx context.Context, p prompt.Prompt) (string, error) {
	messages := make([]chatCompletionMessage, 0, 2)
	if strings.TrimSpace(p.System) != "" {
		messages = append(messages, chatCompletionMessage{Role: "system", Content: p.System})
	}
	messages = append(messages, chatCompletionMessage{Role: "user", Content: p.Question})

	payload, err := json.Marshal(chatCompletionRequest{
		Model:    c.model,
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("marshal %s request: %w", c.provider, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create %s request: %w", c.provider, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call %s chat completion: %w", c.provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("%s chat completion failed: status %d: %s", c.provider, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode %s response: %w", c.provider, err)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("%s response did not include an answer", c.provider)
	}

	return result.Choices[0].Message.Content, nil
}
