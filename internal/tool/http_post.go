package tool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPPostTool struct{ Client *HTTPClient }

func (t HTTPPostTool) Name() string        { return "http_post" }
func (t HTTPPostTool) Description() string { return "发送受控 HTTP POST 请求" }

func (t HTTPPostTool) Execute(ctx context.Context, input string) (string, error) {
	requestInput, err := decodeHTTPRequestInput(input)
	if err != nil {
		return "", err
	}
	return executeHTTPRequest(ctx, t.Client, http.MethodPost, requestInput)
}

func executeHTTPRequest(ctx context.Context, client *HTTPClient, method string, input HTTPRequestInput) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if client == nil || client.client == nil {
		return "", fmt.Errorf("HTTP client is not configured")
	}
	if err := client.allowed(input.URL); err != nil {
		return "", err
	}

	var body io.Reader
	if method == http.MethodPost && len(input.Body) > 0 && string(input.Body) != "null" {
		body = bytes.NewReader(input.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, input.URL, body)
	if err != nil {
		return "", fmt.Errorf("create HTTP request: %w", err)
	}
	for key, value := range input.Headers {
		req.Header.Set(key, value)
	}
	if method == http.MethodPost && body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	data, truncated, err := readHTTPResponse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read HTTP response: %w", err)
	}
	result := fmt.Sprintf("status=%d\nbody=%s", resp.StatusCode, string(data))
	if truncated {
		result += fmt.Sprintf("\n[body truncated at %d bytes]", maxHTTPResponseBytes)
	}
	return result, nil
}

func readHTTPResponse(body io.Reader) ([]byte, bool, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxHTTPResponseBytes+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) <= maxHTTPResponseBytes {
		return data, false, nil
	}
	return data[:maxHTTPResponseBytes], true, nil
}
