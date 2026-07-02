package config

import "os"

// ApplyModeOverride updates provider defaults when envMode overrides fileMode.
func ApplyModeOverride(llm LLMConfig, envMode string) LLMConfig {
	mode := EffectiveLLMMode(llm.Mode, envMode)
	if mode == llm.Mode {
		return llm
	}

	llm.Mode = mode
	llm.Model, llm.BaseURL = DefaultsForLLMMode(mode)
	return llm
}

func EnvLLMMode() string {
	return os.Getenv("LLM_MODE")
}
