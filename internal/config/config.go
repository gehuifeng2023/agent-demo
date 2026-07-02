package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config contains runtime settings loaded at process startup.
type Config struct {
	Service ServiceConfig `yaml:"service"`
	LLM     LLMConfig     `yaml:"llm"`
}

type ServiceConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type LLMConfig struct {
	Mode           string `yaml:"mode"`
	APIKey         string `yaml:"api_key"`
	Model          string `yaml:"model"`
	BaseURL        string `yaml:"base_url"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

// Load reads YAML configuration from path and applies defaults.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

// EffectiveLLMMode returns the selected LLM mode, with envMode taking priority.
func EffectiveLLMMode(fileMode, envMode string) string {
	if strings.TrimSpace(envMode) != "" {
		return normalizeMode(envMode)
	}
	return normalizeMode(fileMode)
}

func normalizeMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}
