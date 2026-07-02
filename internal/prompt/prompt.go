package prompt

import "strings"

// Prompt is the normalized input passed to the LLM layer.
type Prompt struct {
	System   string
	Question string
}

// Build creates a concise assistant prompt for the chat service.
func Build(question string) Prompt {
	return Prompt{
		System:   "你是一个简洁、准确的中文技术助手。",
		Question: strings.TrimSpace(question),
	}
}
