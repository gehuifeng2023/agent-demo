package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAICompatibleBaseURL = "https://api.openai.com/v1"

type OpenAICompatibleClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAICompatibleClientWithConfig(apiKey, model, baseURL string, timeout time.Duration) (*OpenAICompatibleClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("embedding API key is empty")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("embedding model is empty")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultOpenAICompatibleBaseURL
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &OpenAICompatibleClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func (c *OpenAICompatibleClient) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	body, err := json.Marshal(openAIEmbeddingRequest{Model: c.model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embedding API: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read embedding response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("embedding API status=%d body=%s", resp.StatusCode, string(responseBody))
	}

	var response openAIEmbeddingResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("unmarshal embedding response: %w", err)
	}
	return orderedEmbeddings(response.Data, len(texts))
}

func (c *OpenAICompatibleClient) endpoint() string {
	baseURL := strings.TrimRight(c.baseURL, "/")
	if baseURL == "" {
		baseURL = defaultOpenAICompatibleBaseURL
	}
	return baseURL + "/embeddings"
}

func orderedEmbeddings(data []struct {
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}, count int) ([][]float64, error) {
	if len(data) != count {
		return nil, fmt.Errorf("embedding response count=%d, want %d", len(data), count)
	}

	vectors := make([][]float64, count)
	dimension := 0
	for _, item := range data {
		if item.Index < 0 || item.Index >= count {
			return nil, fmt.Errorf("embedding response index=%d is out of range", item.Index)
		}
		if vectors[item.Index] != nil {
			return nil, fmt.Errorf("embedding response contains duplicate index=%d", item.Index)
		}
		if len(item.Embedding) == 0 {
			return nil, fmt.Errorf("embedding response index=%d is empty", item.Index)
		}
		if dimension == 0 {
			dimension = len(item.Embedding)
		} else if len(item.Embedding) != dimension {
			return nil, fmt.Errorf("embedding response index=%d has dimension=%d, want %d", item.Index, len(item.Embedding), dimension)
		}
		vectors[item.Index] = append([]float64(nil), item.Embedding...)
	}
	for index, vector := range vectors {
		if vector == nil {
			return nil, fmt.Errorf("embedding response is missing index=%d", index)
		}
	}
	return vectors, nil
}
