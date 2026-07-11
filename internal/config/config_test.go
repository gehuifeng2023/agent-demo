package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMapsLocalYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "local.yaml")
	data := []byte(`
server:
  addr: ":9090"
llm:
  mode: gemini
  api_key: key
  model: gemini-test
  base_url: http://example.com
  timeout_seconds: 15
rag:
  top_k: 5
upload:
  dir: uploads
  max_size_mb: 7
knowledge:
  root_dir: knowledge
session:
  max_messages: 12
  recent_limit: 4
tool:
  enabled: false
  root_dir: tool-root
  http_allowed_hosts: [api.example.com]
  http_timeout_seconds: 12
workflow:
  definitions:
    - id: service_check
      nodes:
        - name: health
          tool: http_get
          input: '{{question}}'
          output_key: health
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Addr != ":9090" {
		t.Fatalf("expected addr :9090, got %q", cfg.Server.Addr)
	}
	if cfg.LLM.Mode != "gemini" || cfg.LLM.APIKey != "key" || cfg.LLM.Model != "gemini-test" {
		t.Fatalf("unexpected llm config: %#v", cfg.LLM)
	}
	if cfg.LLM.BaseURL != "http://example.com" {
		t.Fatalf("expected base url, got %q", cfg.LLM.BaseURL)
	}
	if cfg.LLMTimeout() != 15*time.Second {
		t.Fatalf("expected 15s timeout, got %s", cfg.LLMTimeout())
	}
	if cfg.RAG.TopK != 5 {
		t.Fatalf("expected top_k 5, got %d", cfg.RAG.TopK)
	}
	if cfg.Upload.Dir != "uploads" || cfg.UploadMaxBytes() != 7<<20 {
		t.Fatalf("unexpected upload config: %#v bytes=%d", cfg.Upload, cfg.UploadMaxBytes())
	}
	if cfg.Knowledge.RootDir != "knowledge" {
		t.Fatalf("expected knowledge root, got %q", cfg.Knowledge.RootDir)
	}
	if cfg.Session.MaxMessages != 12 || cfg.Session.RecentLimit != 4 {
		t.Fatalf("unexpected session config: %#v", cfg.Session)
	}
	if cfg.ToolEnabled() {
		t.Fatal("expected tool to be disabled")
	}
	if cfg.ToolRootDir() != "tool-root" {
		t.Fatalf("expected tool root, got %q", cfg.ToolRootDir())
	}
	if len(cfg.Tool.HTTPAllowedHosts) != 1 || cfg.Tool.HTTPAllowedHosts[0] != "api.example.com" {
		t.Fatalf("unexpected HTTP allow hosts %#v", cfg.Tool.HTTPAllowedHosts)
	}
	if cfg.HTTPToolTimeout() != 12*time.Second {
		t.Fatalf("expected HTTP timeout 12s, got %s", cfg.HTTPToolTimeout())
	}
	if len(cfg.Workflow.Definitions) != 1 || cfg.Workflow.Definitions[0].ID != "service_check" {
		t.Fatalf("unexpected workflow definitions %#v", cfg.Workflow.Definitions)
	}
	node := cfg.Workflow.Definitions[0].Nodes[0]
	if node.Name != "health" || node.Tool != "http_get" || node.Input != "{{question}}" || node.OutputKey != "health" {
		t.Fatalf("unexpected workflow node %#v", node)
	}
}

func TestApplyDefaultsUsesCurrentHardCodedValues(t *testing.T) {
	cfg := Config{}

	cfg.ApplyDefaults()

	if cfg.Server.Addr != DefaultServerAddr {
		t.Fatalf("expected default addr %q, got %q", DefaultServerAddr, cfg.Server.Addr)
	}
	if cfg.LLM.Mode != DefaultLLMMode {
		t.Fatalf("expected default llm mode %q, got %q", DefaultLLMMode, cfg.LLM.Mode)
	}
	if cfg.LLMTimeout() != DefaultLLMTimeoutSeconds*time.Second {
		t.Fatalf("expected default llm timeout, got %s", cfg.LLMTimeout())
	}
	if cfg.RAG.TopK != DefaultRAGTopK {
		t.Fatalf("expected default top_k %d, got %d", DefaultRAGTopK, cfg.RAG.TopK)
	}
	if cfg.Upload.Dir != DefaultUploadDir || cfg.UploadMaxBytes() != DefaultUploadMaxSizeMB<<20 {
		t.Fatalf("unexpected default upload config: %#v bytes=%d", cfg.Upload, cfg.UploadMaxBytes())
	}
	if cfg.Knowledge.RootDir != DefaultKnowledgeRootDir {
		t.Fatalf("expected default knowledge root %q, got %q", DefaultKnowledgeRootDir, cfg.Knowledge.RootDir)
	}
	if cfg.Session.MaxMessages != DefaultSessionMaxMessages || cfg.Session.RecentLimit != DefaultSessionRecentLimit {
		t.Fatalf("unexpected default session config: %#v", cfg.Session)
	}
	if cfg.ToolEnabled() != DefaultToolEnabled {
		t.Fatalf("expected default tool enabled %v, got %v", DefaultToolEnabled, cfg.ToolEnabled())
	}
	if cfg.ToolRootDir() != DefaultKnowledgeRootDir {
		t.Fatalf("expected default tool root %q, got %q", DefaultKnowledgeRootDir, cfg.ToolRootDir())
	}
	if cfg.HTTPToolTimeout() != DefaultHTTPTimeoutSeconds*time.Second {
		t.Fatalf("expected default HTTP timeout %ds, got %s", DefaultHTTPTimeoutSeconds, cfg.HTTPToolTimeout())
	}
}
