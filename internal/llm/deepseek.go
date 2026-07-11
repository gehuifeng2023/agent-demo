package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultDeepSeekBaseURL = "https://api.deepseek.com"

type DeepSeekClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewDeepSeekClient() (*DeepSeekClient, error) {
	return NewDeepSeekClientWithConfig(
		os.Getenv("DEEPSEEK_API_KEY"),
		os.Getenv("LLM_MODEL"),
		"",
		60*time.Second,
	)
}

func NewDeepSeekClientWithConfig(apiKey, model, baseURL string, timeout time.Duration) (*DeepSeekClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY is empty")
	}
	if strings.TrimSpace(model) == "" {
		model = "deepseek-chat"
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultDeepSeekBaseURL
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &DeepSeekClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []deepSeekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
	} `json:"choices"`
}

func (c *DeepSeekClient) Generate(ctx context.Context, prompt string) (string, error) {
	body, err := c.doRequest(ctx, prompt, false)
	if err != nil {
		return "", err
	}

	var result deepSeekResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal DeepSeek response: %w", err)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("DeepSeek response content is empty")
	}
	return result.Choices[0].Message.Content, nil
}

func (c *DeepSeekClient) Stream(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	parts := make(chan string)
	errs := make(chan error, 1)

	go func() {
		defer close(parts)
		defer close(errs)

		body, err := c.requestBody(prompt, true)
		if err != nil {
			errs <- err
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("create DeepSeek request: %w", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errs <- fmt.Errorf("call DeepSeek stream API: %w", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			responseBody, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				errs <- fmt.Errorf("read DeepSeek error response: %w", readErr)
				return
			}
			errs <- fmt.Errorf("DeepSeek API status=%d body=%s", resp.StatusCode, string(responseBody))
			return
		}

		if err := readDeepSeekSSE(ctx, resp.Body, parts); err != nil {
			errs <- err
		}
	}()

	return parts, errs
}

func (c *DeepSeekClient) doRequest(ctx context.Context, prompt string, stream bool) ([]byte, error) {
	body, err := c.requestBody(prompt, stream)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create DeepSeek request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call DeepSeek API: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read DeepSeek response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("DeepSeek API status=%d body=%s", resp.StatusCode, string(responseBody))
	}
	return responseBody, nil
}

func (c *DeepSeekClient) requestBody(prompt string, stream bool) ([]byte, error) {
	body, err := json.Marshal(deepSeekRequest{
		Model:    c.model,
		Messages: []deepSeekMessage{{Role: "user", Content: prompt}},
		Stream:   stream,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal DeepSeek request: %w", err)
	}
	return body, nil
}

func (c *DeepSeekClient) endpoint() string {
	baseURL := strings.TrimRight(c.baseURL, "/")
	if baseURL == "" {
		baseURL = defaultDeepSeekBaseURL
	}
	return baseURL + "/chat/completions"
}

func readDeepSeekSSE(ctx context.Context, body io.Reader, parts chan<- string) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return nil
		}

		var chunk deepSeekResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("unmarshal DeepSeek stream chunk: %w", err)
		}
		if len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" {
			continue
		}
		select {
		case parts <- chunk.Choices[0].Delta.Content:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read DeepSeek stream: %w", err)
	}
	return nil
}
