package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-demo/internal/config"
	"agent-demo/internal/embedding"
	"agent-demo/internal/llm"
)

func TestNewLLMClientFromConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	t.Setenv("DEEPSEEK_API_KEY", "deepseek-key")

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

	t.Run("deepseek", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.LLM.Mode = "deepseek"

		client, mode, err := newLLMClientFromConfig(cfg)
		if err != nil {
			t.Fatalf("newLLMClientFromConfig failed: %v", err)
		}
		if mode != "deepseek" {
			t.Fatalf("expected deepseek mode, got %q", mode)
		}
		if _, ok := client.(*llm.DeepSeekClient); !ok {
			t.Fatalf("expected *llm.DeepSeekClient, got %T", client)
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

func TestNewEmbeddingClientFromConfig(t *testing.T) {
	t.Setenv("EMBEDDING_API_KEY", "embedding-key")

	t.Run("disabled", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()

		client, err := newEmbeddingClientFromConfig(cfg)
		if err != nil {
			t.Fatalf("newEmbeddingClientFromConfig failed: %v", err)
		}
		if client != nil {
			t.Fatalf("expected nil client, got %T", client)
		}
	})

	t.Run("openai compatible", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.Embedding.Mode = "openai_compatible"
		cfg.Embedding.Model = "test-embedding-model"

		client, err := newEmbeddingClientFromConfig(cfg)
		if err != nil {
			t.Fatalf("newEmbeddingClientFromConfig failed: %v", err)
		}
		if _, ok := client.(*embedding.OpenAICompatibleClient); !ok {
			t.Fatalf("expected *embedding.OpenAICompatibleClient, got %T", client)
		}
	})

	t.Run("requires model", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.ApplyDefaults()
		cfg.Embedding.Mode = "openai_compatible"

		if _, err := newEmbeddingClientFromConfig(cfg); err == nil {
			t.Fatal("expected model error")
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

	unifiedRetriever, err := newRetrieverFromDefaultKnowledge(dir, nil)
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

	logAnalyzer, ok := registry.Get("log_analyzer")
	if !ok {
		t.Fatal("expected log_analyzer")
	}
	logResult, err := logAnalyzer.Execute(context.Background(), "request_id=abc status=502")
	if err != nil {
		t.Fatalf("execute log_analyzer: %v", err)
	}
	if !strings.Contains(logResult, `"error_type": "gateway_502"`) {
		t.Fatalf("unexpected log analysis result %q", logResult)
	}

	if _, ok := registry.Get("http_get"); ok {
		t.Fatal("expected HTTP tool to remain disabled without allowlisted hosts")
	}
	cfg.Tool.HTTPAllowedHosts = []string{"api.example.com"}
	registry = newToolRegistry(cfg)
	if _, ok := registry.Get("http_get"); !ok {
		t.Fatal("expected http_get for configured allowlisted hosts")
	}
	if _, ok := registry.Get("http_post"); !ok {
		t.Fatal("expected http_post for configured allowlisted hosts")
	}

	disabled := false
	cfg.Tool.Enabled = &disabled
	if registry := newToolRegistry(cfg); registry != nil {
		t.Fatal("expected nil registry when tool is disabled")
	}
}

func TestNewWorkflowRegistry(t *testing.T) {
	cfg := &config.Config{}
	cfg.ApplyDefaults()
	cfg.Workflow.Definitions = []config.WorkflowDefinitionConfig{{
		ID: "analyze",
		Nodes: []config.WorkflowNodeConfig{{
			Name:      "log",
			Tool:      "log_analyzer",
			Input:     "{{question}}",
			OutputKey: "analysis",
		}},
	}}

	registry, err := newWorkflowRegistry(cfg, newToolRegistry(cfg))
	if err != nil {
		t.Fatalf("new workflow registry: %v", err)
	}
	if _, ok := registry.Get("analyze"); !ok {
		t.Fatal("expected configured workflow")
	}

	cfg.Workflow.Definitions[0].Nodes[0].Tool = "missing"
	if _, err := newWorkflowRegistry(cfg, newToolRegistry(cfg)); err == nil || !strings.Contains(err.Error(), "tool not found") {
		t.Fatalf("expected missing tool error, got %v", err)
	}
}
