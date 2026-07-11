package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeepSeekClientGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"deepseek answer"}}]}`))
	}))
	defer server.Close()

	client, err := NewDeepSeekClientWithConfig("test-key", "deepseek-chat", server.URL, 0)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	got, err := client.Generate(context.Background(), "hello")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if got != "deepseek answer" {
		t.Fatalf("unexpected answer: %q", got)
	}
}

func TestDeepSeekClientStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"hidden\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	client, err := NewDeepSeekClientWithConfig("test-key", "deepseek-chat", server.URL, 0)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	parts, errs := client.Stream(context.Background(), "hello")
	var got strings.Builder
	for part := range parts {
		got.WriteString(part)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream failed: %v", err)
		}
	}
	if got.String() != "hello world" {
		t.Fatalf("unexpected stream output: %q", got.String())
	}
}

func TestNewDeepSeekClientRequiresAPIKey(t *testing.T) {
	_, err := NewDeepSeekClientWithConfig("", "", "", 0)
	if err == nil {
		t.Fatal("expected API key error")
	}
}
