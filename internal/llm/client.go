package llm

import (
	"context"
	"strings"

	"agent-demo/internal/prompt"
)

// Client generates an answer for a prompt.
type Client interface {
	Generate(ctx context.Context, p prompt.Prompt) (string, error)
}

// MockClient is a deterministic local LLM substitute for the demo.
type MockClient struct{}

// NewMockClient creates a local mock LLM client.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Generate returns a stable answer without calling an external provider.
func (c *MockClient) Generate(ctx context.Context, p prompt.Prompt) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	question := strings.ToLower(p.Question)
	if strings.Contains(question, "apisix") {
		return "APISIX 是一个云原生 API 网关，可用于流量管理、服务代理、认证鉴权、限流、观测和插件扩展等场景。", nil
	}

	return "这是一个本地模拟的 Demo 回答。当前服务已收到你的问题，并通过 prompt、service 和 llm 分层生成响应。", nil
}
