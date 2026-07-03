package prompt

import (
	"fmt"
	"strings"
)

// DefaultBuilder builds prompts from the built-in template set.
type DefaultBuilder struct{}

// NewDefaultBuilder creates a builder with the built-in prompt templates.
func NewDefaultBuilder() *DefaultBuilder {
	return &DefaultBuilder{}
}

// Build creates a prompt from a typed request.
func (b *DefaultBuilder) Build(req Request) (Prompt, error) {
	promptType := req.Type
	if promptType == "" {
		promptType = TypeQA
	}

	tmpl, err := newTemplate(promptType)
	if err != nil {
		return Prompt{}, err
	}

	req.Input = strings.TrimSpace(req.Input)
	req.Context = strings.TrimSpace(req.Context)
	return tmpl.Build(req), nil
}

// Build creates a concise QA prompt for existing callers.
func Build(question string) Prompt {
	p, _ := NewDefaultBuilder().Build(Request{
		Type:  TypeQA,
		Input: question,
	})
	return p
}

func newTemplate(promptType Type) (template, error) {
	switch promptType {
	case TypeQA:
		return chatTemplate{}, nil
	case TypeRAGQA:
		return ragChatTemplate{}, nil
	case TypeLogAnalysis:
		return logAnalysisTemplate{}, nil
	case TypeSummary:
		return summarizeTemplate{}, nil
	case TypeTaskBreakdown:
		return taskBreakdownTemplate{}, nil
	case TypeReport:
		return reportTemplate{}, nil
	default:
		return nil, fmt.Errorf("unsupported prompt type %q", promptType)
	}
}
