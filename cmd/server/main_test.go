package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-demo/internal/config"
	"agent-demo/internal/llm"
)

func TestNewLLMClientFromConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("GEMINI_API_KEY", "gemini-key")

	t.Run("mock", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.LLM.Mode = "mock"

		client, mode, err := newLLMClientFromConfig(cfg)
		if err != nil {
			t.Fatalf("newLLMClientFromConfig failed: %v", err)
		}
		if mode != "mock" {
			t.Fatalf("expected mock mode, got %q", mode)
		}
		if _, ok := client.(*llm.MockClient); !ok {
			t.Fatalf("expected *llm.MockClient, got %T", client)
		}
	})

	t.Run("openai default", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.LLM.Mode = "openai"

		client, mode, err := newLLMClientFromConfig(cfg)
		if err != nil {
			t.Fatalf("newLLMClientFromConfig failed: %v", err)
		}
		if mode != "openai" {
			t.Fatalf("expected openai mode, got %q", mode)
		}
		if _, ok := client.(*llm.OpenAIClient); !ok {
			t.Fatalf("expected *llm.OpenAIClient, got %T", client)
		}
	})

	t.Run("gemini", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.LLM.Mode = "gemini"

		client, mode, err := newLLMClientFromConfig(cfg)
		if err != nil {
			t.Fatalf("newLLMClientFromConfig failed: %v", err)
		}
		if mode != "gemini" {
			t.Fatalf("expected gemini mode, got %q", mode)
		}
		if _, ok := client.(*llm.GeminiClient); !ok {
			t.Fatalf("expected *llm.GeminiClient, got %T", client)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.LLM.Mode = "other"

		_, _, err := newLLMClientFromConfig(cfg)
		if err == nil {
			t.Fatal("expected error for unsupported mode")
		}
	})
}

func TestLoadConfigUsesConfigPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "local.yaml")
	if err := os.WriteFile(path, []byte("server:\n  addr: \":9090\"\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CONFIG_PATH", path)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Server.Addr != ":9090" {
		t.Fatalf("expected configured addr, got %q", cfg.Server.Addr)
	}
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

func TestNewToolRegistry(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "faq.md"), []byte("tool config content"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg := &config.Config{}
	cfg.ApplyDefaults()
	cfg.Tool.RootDir = root

	registry := newToolRegistry(cfg)
	if registry == nil {
		t.Fatal("expected registry")
	}

	fileReader, ok := registry.Get("file_reader")
	if !ok {
		t.Fatal("expected file_reader")
	}
	got, err := fileReader.Execute(context.Background(), "faq.md")
	if err != nil {
		t.Fatalf("execute file_reader: %v", err)
	}
	if got != "tool config content" {
		t.Fatalf("unexpected content %q", got)
	}

	disabled := false
	cfg.Tool.Enabled = &disabled
	if registry := newToolRegistry(cfg); registry != nil {
		t.Fatal("expected nil registry when tool is disabled")
	}
}
