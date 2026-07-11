package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatibleClientEmbedUsesAPIContractAndInputOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("expected /v1/embeddings, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected authorization header: %q", got)
		}
		var body openAIEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body.Model != "test-model" || len(body.Input) != 2 || body.Input[0] != "first" || body.Input[1] != "second" {
			t.Errorf("unexpected request body: %#v", body)
		}
		_, _ = w.Write([]byte(`{"data":[{"index":1,"embedding":[3,4]},{"index":0,"embedding":[1,2]}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClientWithConfig("test-key", "test-model", server.URL+"/v1", 0)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	vectors, err := client.Embed(context.Background(), []string{"first", "second"})
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if len(vectors) != 2 || vectors[0][0] != 1 || vectors[1][0] != 3 {
		t.Fatalf("unexpected vectors: %#v", vectors)
	}
}

func TestOpenAICompatibleClientEmbedRejectsInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"index":0,"embedding":[1,2]},{"index":1,"embedding":[3]}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClientWithConfig("test-key", "test-model", server.URL, 0)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := client.Embed(context.Background(), []string{"first", "second"}); err == nil {
		t.Fatal("expected invalid response error")
	}
}

func TestOpenAICompatibleClientEmbedReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClientWithConfig("test-key", "test-model", server.URL, 0)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if _, err := client.Embed(context.Background(), []string{"first"}); err == nil {
		t.Fatal("expected API error")
	}
}

func TestOpenAICompatibleClientEmbedHonorsCanceledContext(t *testing.T) {
	client, err := NewOpenAICompatibleClientWithConfig("test-key", "test-model", "http://127.0.0.1:1/v1", 0)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Embed(ctx, []string{"first"}); err == nil {
		t.Fatal("expected canceled context error")
	}
}
