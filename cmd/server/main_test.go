package main

import (
	"path/filepath"
	"strings"
	"testing"

	"agent-demo/internal/llm"
)

func TestNewLLMClientFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("GEMINI_API_KEY", "gemini-key")

	t.Run("mock", func(t *testing.T) {
		t.Setenv("LLM_MODE", "mock")

		client, mode, err := newLLMClientFromEnv()
		if err != nil {
			t.Fatalf("newLLMClientFromEnv failed: %v", err)
		}
		if mode != "mock" {
			t.Fatalf("expected mock mode, got %q", mode)
		}
		if _, ok := client.(*llm.MockClient); !ok {
			t.Fatalf("expected *llm.MockClient, got %T", client)
		}
	})

	t.Run("openai default", func(t *testing.T) {
		t.Setenv("LLM_MODE", "")

		client, mode, err := newLLMClientFromEnv()
		if err != nil {
			t.Fatalf("newLLMClientFromEnv failed: %v", err)
		}
		if mode != "openai" {
			t.Fatalf("expected openai mode, got %q", mode)
		}
		if _, ok := client.(*llm.OpenAIClient); !ok {
			t.Fatalf("expected *llm.OpenAIClient, got %T", client)
		}
	})

	t.Run("gemini", func(t *testing.T) {
		t.Setenv("LLM_MODE", "gemini")

		client, mode, err := newLLMClientFromEnv()
		if err != nil {
			t.Fatalf("newLLMClientFromEnv failed: %v", err)
		}
		if mode != "gemini" {
			t.Fatalf("expected gemini mode, got %q", mode)
		}
		if _, ok := client.(*llm.GeminiClient); !ok {
			t.Fatalf("expected *llm.GeminiClient, got %T", client)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		t.Setenv("LLM_MODE", "other")

		_, _, err := newLLMClientFromEnv()
		if err == nil {
			t.Fatal("expected error for unsupported mode")
		}
	})
}

func TestNewRetrieverFromDefaultKnowledge(t *testing.T) {
	dir := filepath.Join("..", "..", "knowledge_attachment", "default")

	unifiedRetriever, err := newRetrieverFromDefaultKnowledge(dir)
	if err != nil {
		t.Fatalf("newRetrieverFromDefaultKnowledge failed: %v", err)
	}

	chunks := unifiedRetriever.Retrieve("RAG", []string{"default"}, nil, 3)
	if len(chunks) == 0 {
		t.Fatal("expected default knowledge chunks")
	}
	if !strings.Contains(chunks[0].Source, "faq.md") {
		t.Fatalf("expected faq source, got %q", chunks[0].Source)
	}
}
