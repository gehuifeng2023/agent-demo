package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agent-demo/internal/prompt"
)

func TestMockClientGenerateAPISIXAnswer(t *testing.T) {
	client := NewMockClient()

	answer, err := client.Generate(context.Background(), prompt.Build("什么是 APISIX？"))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(answer, "APISIX") {
		t.Fatalf("Generate() answer = %q, want APISIX content", answer)
	}
}

func TestBailianClientGenerateSendsChatCompletionRequest(t *testing.T) {
	var gotAuth string
	var gotRequest bailianRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s, want /chat/completions", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"真实模型回答"}}]}`))
	}))
	defer server.Close()

	client, err := NewBailianClient("test-key", "qwen-plus", server.URL, time.Second)
	if err != nil {
		t.Fatalf("NewBailianClient() error = %v", err)
	}

	answer, err := client.Generate(context.Background(), prompt.Build("你好"))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if answer != "真实模型回答" {
		t.Fatalf("Generate() answer = %q, want 真实模型回答", answer)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization = %q, want Bearer test-key", gotAuth)
	}
	if gotRequest.Model != "qwen-plus" {
		t.Fatalf("model = %q, want qwen-plus", gotRequest.Model)
	}
	if len(gotRequest.Messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(gotRequest.Messages))
	}
	if gotRequest.Messages[0].Role != "system" || gotRequest.Messages[1].Role != "user" {
		t.Fatalf("messages roles = %#v, want system then user", gotRequest.Messages)
	}
	if gotRequest.Messages[1].Content != "你好" {
		t.Fatalf("user content = %q, want 你好", gotRequest.Messages[1].Content)
	}
}

func TestBailianClientGenerateReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewBailianClient("test-key", "qwen-plus", server.URL, time.Second)
	if err != nil {
		t.Fatalf("NewBailianClient() error = %v", err)
	}

	_, err = client.Generate(context.Background(), prompt.Build("你好"))
	if err == nil || !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("Generate() error = %v, want status 400", err)
	}
}

func TestDeepSeekClientGenerateSendsChatCompletionRequest(t *testing.T) {
	var gotAuth string
	var gotRequest chatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s, want /chat/completions", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"deepseek answer"}}]}`))
	}))
	defer server.Close()

	client, err := NewDeepSeekClient("test-key", "deepseek-v4-flash", server.URL, time.Second)
	if err != nil {
		t.Fatalf("NewDeepSeekClient() error = %v", err)
	}

	answer, err := client.Generate(context.Background(), prompt.Build("你好"))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if answer != "deepseek answer" {
		t.Fatalf("Generate() answer = %q, want deepseek answer", answer)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization = %q, want Bearer test-key", gotAuth)
	}
	if gotRequest.Model != "deepseek-v4-flash" {
		t.Fatalf("model = %q, want deepseek-v4-flash", gotRequest.Model)
	}
}

func TestDeepSeekClientGenerateReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewDeepSeekClient("test-key", "deepseek-v4-flash", server.URL, time.Second)
	if err != nil {
		t.Fatalf("NewDeepSeekClient() error = %v", err)
	}

	_, err = client.Generate(context.Background(), prompt.Build("你好"))
	if err == nil || !strings.Contains(err.Error(), "deepseek chat completion failed: status 400") {
		t.Fatalf("Generate() error = %v, want deepseek status 400", err)
	}
}

func TestGeminiClientGenerateSendsChatCompletionRequest(t *testing.T) {
	var gotAuth string
	var gotRequest chatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s, want /chat/completions", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"gemini answer"}}]}`))
	}))
	defer server.Close()

	client, err := NewGeminiClient("test-key", "gemini-3.5-flash", server.URL, time.Second)
	if err != nil {
		t.Fatalf("NewGeminiClient() error = %v", err)
	}

	answer, err := client.Generate(context.Background(), prompt.Build("你好"))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if answer != "gemini answer" {
		t.Fatalf("Generate() answer = %q, want gemini answer", answer)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization = %q, want Bearer test-key", gotAuth)
	}
	if gotRequest.Model != "gemini-3.5-flash" {
		t.Fatalf("model = %q, want gemini-3.5-flash", gotRequest.Model)
	}
}

func TestGeminiClientGenerateReturnsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewGeminiClient("test-key", "gemini-3.5-flash", server.URL, time.Second)
	if err != nil {
		t.Fatalf("NewGeminiClient() error = %v", err)
	}

	_, err = client.Generate(context.Background(), prompt.Build("你好"))
	if err == nil || !strings.Contains(err.Error(), "gemini chat completion failed: status 400") {
		t.Fatalf("Generate() error = %v, want gemini status 400", err)
	}
}
