package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"agent-demo/internal/llm"
	"agent-demo/internal/prompt"
)

var ErrQuestionRequired = errors.New("question is required")

// ChatService handles chat business logic.
type ChatService struct {
	llm     llm.Client
	builder prompt.Builder
}

// NewChatService creates a chat service with its LLM dependency.
func NewChatService(client llm.Client) *ChatService {
	return NewChatServiceWithPromptBuilder(client, prompt.NewDefaultBuilder())
}

// NewChatServiceWithPromptBuilder creates a chat service with explicit dependencies.
func NewChatServiceWithPromptBuilder(client llm.Client, builder prompt.Builder) *ChatService {
	return &ChatService{
		llm:     client,
		builder: builder,
	}
}

// Ask validates the question, builds a prompt, and generates an answer.
func (s *ChatService) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", ErrQuestionRequired
	}

	builtPrompt, err := s.builder.Build(prompt.Request{
		Type:  prompt.TypeQA,
		Input: question,
	})
	if err != nil {
		return "", fmt.Errorf("build prompt: %w", err)
	}

	return s.llm.Generate(ctx, builtPrompt)
}
