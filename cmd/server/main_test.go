package main

import (
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
