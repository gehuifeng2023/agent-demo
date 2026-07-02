package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesDefaultsAndReadsLLMConfig(t *testing.T) {
	path := writeConfig(t, `
service:
  name: agent-demo
llm:
  mode: model
  api_key: test-key
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Service.Address != DefaultServiceAddress {
		t.Fatalf("Service.Address = %q, want %q", cfg.Service.Address, DefaultServiceAddress)
	}
	if cfg.LLM.Mode != "model" {
		t.Fatalf("LLM.Mode = %q, want model", cfg.LLM.Mode)
	}
	if cfg.LLM.APIKey != "test-key" {
		t.Fatalf("LLM.APIKey = %q, want test-key", cfg.LLM.APIKey)
	}
	if cfg.LLM.Model != DefaultLLMModel {
		t.Fatalf("LLM.Model = %q, want %q", cfg.LLM.Model, DefaultLLMModel)
	}
	if cfg.LLM.BaseURL != DefaultLLMBaseURL {
		t.Fatalf("LLM.BaseURL = %q, want %q", cfg.LLM.BaseURL, DefaultLLMBaseURL)
	}
	if cfg.LLM.TimeoutSeconds != DefaultLLMTimeout {
		t.Fatalf("LLM.TimeoutSeconds = %d, want %d", cfg.LLM.TimeoutSeconds, DefaultLLMTimeout)
	}
}

func TestEffectiveLLMModePrefersEnv(t *testing.T) {
	got := EffectiveLLMMode("mock", " MODEL ")
	if got != "model" {
		t.Fatalf("EffectiveLLMMode() = %q, want model", got)
	}
}

func TestLoadAppliesDeepSeekDefaults(t *testing.T) {
	path := writeConfig(t, `
llm:
  mode: deepseek
  api_key: test-key
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LLM.Mode != "deepseek" {
		t.Fatalf("LLM.Mode = %q, want deepseek", cfg.LLM.Mode)
	}
	if cfg.LLM.Model != DefaultDeepSeekModel {
		t.Fatalf("LLM.Model = %q, want %q", cfg.LLM.Model, DefaultDeepSeekModel)
	}
	if cfg.LLM.BaseURL != DefaultDeepSeekBaseURL {
		t.Fatalf("LLM.BaseURL = %q, want %q", cfg.LLM.BaseURL, DefaultDeepSeekBaseURL)
	}
}

func TestLoadAppliesGeminiDefaults(t *testing.T) {
	path := writeConfig(t, `
llm:
  mode: gemini
  api_key: test-key
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LLM.Mode != "gemini" {
		t.Fatalf("LLM.Mode = %q, want gemini", cfg.LLM.Mode)
	}
	if cfg.LLM.Model != DefaultGeminiModel {
		t.Fatalf("LLM.Model = %q, want %q", cfg.LLM.Model, DefaultGeminiModel)
	}
	if cfg.LLM.BaseURL != DefaultGeminiBaseURL {
		t.Fatalf("LLM.BaseURL = %q, want %q", cfg.LLM.BaseURL, DefaultGeminiBaseURL)
	}
}

func TestApplyModeOverrideUpdatesProviderDefaults(t *testing.T) {
	llm := LLMConfig{
		Mode:    "deepseek",
		APIKey:  "test-key",
		Model:   DefaultDeepSeekModel,
		BaseURL: DefaultDeepSeekBaseURL,
	}

	got := ApplyModeOverride(llm, "gemini")

	if got.Mode != "gemini" {
		t.Fatalf("Mode = %q, want gemini", got.Mode)
	}
	if got.Model != DefaultGeminiModel {
		t.Fatalf("Model = %q, want %q", got.Model, DefaultGeminiModel)
	}
	if got.BaseURL != DefaultGeminiBaseURL {
		t.Fatalf("BaseURL = %q, want %q", got.BaseURL, DefaultGeminiBaseURL)
	}
	if got.APIKey != "test-key" {
		t.Fatalf("APIKey = %q, want test-key", got.APIKey)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
