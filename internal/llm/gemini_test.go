package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeminiClientGenerate(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotAPIKey string
	var gotBody responsesRequest
	var decodeErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("x-goog-api-key")

		decodeErr = json.NewDecoder(r.Body).Decode(&gotBody)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output_text":"gemini answer"}`))
	}))
	defer server.Close()

	client := &GeminiClient{
		apiKey:     "test-key",
		model:      "gemini-3.5-flash",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	got, err := client.Generate(context.Background(), "hello")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if got != "gemini answer" {
		t.Fatalf("expected gemini answer, got %q", got)
	}
	if decodeErr != nil {
		t.Fatalf("decode request: %v", decodeErr)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/interactions" {
		t.Fatalf("expected /interactions, got %s", gotPath)
	}
	if gotAPIKey != "test-key" {
		t.Fatalf("expected x-goog-api-key header, got %q", gotAPIKey)
	}
	if gotBody.Model != "gemini-3.5-flash" {
		t.Fatalf("expected model gemini-3.5-flash, got %q", gotBody.Model)
	}
	if gotBody.Input != "hello" {
		t.Fatalf("expected prompt hello, got %q", gotBody.Input)
	}
}

func TestGeminiClientGenerateRejectsEmptyOutputText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output_text":""}`))
	}))
	defer server.Close()

	client := &GeminiClient{
		apiKey:     "test-key",
		model:      "gemini-3.5-flash",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	_, err := client.Generate(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for empty output_text")
	}
}
