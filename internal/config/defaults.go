package config

import (
	"strings"
	"time"
)

const (
	DefaultServiceAddress = ":8080"
	DefaultLLMMode        = "mock"
	DefaultLLMModel       = "qwen-plus"
	DefaultLLMBaseURL     = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	DefaultLLMTimeout     = 30

	DefaultDeepSeekModel   = "deepseek-v4-flash"
	DefaultDeepSeekBaseURL = "https://api.deepseek.com"

	DefaultGeminiModel   = "gemini-3.5-flash"
	DefaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
)

// ApplyDefaults fills unset values with safe local defaults.
func (c *Config) ApplyDefaults() {
	if strings.TrimSpace(c.Service.Address) == "" {
		c.Service.Address = DefaultServiceAddress
	}
	if strings.TrimSpace(c.LLM.Mode) == "" {
		c.LLM.Mode = DefaultLLMMode
	} else {
		c.LLM.Mode = normalizeMode(c.LLM.Mode)
	}

	defaultModel, defaultBaseURL := DefaultsForLLMMode(c.LLM.Mode)
	if strings.TrimSpace(c.LLM.Model) == "" {
		c.LLM.Model = defaultModel
	}
	if strings.TrimSpace(c.LLM.BaseURL) == "" {
		c.LLM.BaseURL = defaultBaseURL
	}
	if c.LLM.TimeoutSeconds <= 0 {
		c.LLM.TimeoutSeconds = DefaultLLMTimeout
	}
}

// DefaultsForLLMMode returns the provider defaults for a normalized LLM mode.
func DefaultsForLLMMode(mode string) (model string, baseURL string) {
	switch normalizeMode(mode) {
	case "deepseek":
		return DefaultDeepSeekModel, DefaultDeepSeekBaseURL
	case "gemini":
		return DefaultGeminiModel, DefaultGeminiBaseURL
	default:
		return DefaultLLMModel, DefaultLLMBaseURL
	}
}

func (c LLMConfig) Timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}
