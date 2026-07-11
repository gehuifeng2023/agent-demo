package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPRequestInput struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body,omitempty"`
}

type HTTPGetTool struct {
	Client *HTTPClient
}

func (t HTTPGetTool) Name() string        { return "http_get" }
func (t HTTPGetTool) Description() string { return "发送受控 HTTP GET 请求" }

func (t HTTPGetTool) Execute(ctx context.Context, input string) (string, error) {
	requestInput, err := decodeHTTPRequestInput(input)
	if err != nil {
		return "", err
	}
	return executeHTTPRequest(ctx, t.Client, http.MethodGet, requestInput)
}

func decodeHTTPRequestInput(input string) (HTTPRequestInput, error) {
	var requestInput HTTPRequestInput
	if err := json.Unmarshal([]byte(input), &requestInput); err != nil {
		return HTTPRequestInput{}, fmt.Errorf("parse HTTP tool input: %w", err)
	}
	if requestInput.URL == "" {
		return HTTPRequestInput{}, fmt.Errorf("HTTP request URL is empty")
	}
	return requestInput, nil
}
