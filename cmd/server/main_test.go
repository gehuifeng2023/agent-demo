package main

import (
	"testing"

	"agent-demo/internal/config"
	"agent-demo/internal/llm"
)

func TestNewLLMClientDefaultsToMock(t *testing.T) {
	cfg := config.Config{}
	cfg.ApplyDefaults()

	client, err := newLLMClient(cfg)
	if err != nil {
		t.Fatalf("newLLMClient() error = %v", err)
	}
	if client == nil {
		t.Fatalf("newLLMClient() returned nil client")
	}
}

func TestNewLLMClientModelRequiresAPIKey(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Mode: "model",
		},
	}
	cfg.ApplyDefaults()

	_, err := newLLMClient(cfg)
	if err == nil {
		t.Fatalf("newLLMClient() error = nil, want api key error")
	}
}

func TestNewLLMClientDeepSeekRequiresAPIKey(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Mode: "deepseek",
		},
	}
	cfg.ApplyDefaults()

	_, err := newLLMClient(cfg)
	if err == nil {
		t.Fatalf("newLLMClient() error = nil, want api key error")
	}
}

func TestNewLLMClientCreatesDeepSeekClient(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Mode:   "deepseek",
			APIKey: "test-key",
		},
	}
	cfg.ApplyDefaults()

	client, err := newLLMClient(cfg)
	if err != nil {
		t.Fatalf("newLLMClient() error = %v", err)
	}
	if _, ok := client.(*llm.DeepSeekClient); !ok {
		t.Fatalf("newLLMClient() type = %T, want *llm.DeepSeekClient", client)
	}
}

func TestNewLLMClientGeminiRequiresAPIKey(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Mode: "gemini",
		},
	}
	cfg.ApplyDefaults()

	_, err := newLLMClient(cfg)
	if err == nil {
		t.Fatalf("newLLMClient() error = nil, want api key error")
	}
}

func TestNewLLMClientCreatesGeminiClient(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Mode:   "gemini",
			APIKey: "test-key",
		},
	}
	cfg.ApplyDefaults()

	client, err := newLLMClient(cfg)
	if err != nil {
		t.Fatalf("newLLMClient() error = %v", err)
	}
	if _, ok := client.(*llm.GeminiClient); !ok {
		t.Fatalf("newLLMClient() type = %T, want *llm.GeminiClient", client)
	}
}
