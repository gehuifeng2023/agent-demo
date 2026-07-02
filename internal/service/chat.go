package service

import (
	"context"
	"errors"
	"strings"

	"agent-demo/internal/llm"
	"agent-demo/internal/prompt"
)

var ErrQuestionRequired = errors.New("question is required")

// ChatService handles chat business logic.
type ChatService struct {
	llm llm.Client
}

// NewChatService creates a chat service with its LLM dependency.
func NewChatService(client llm.Client) *ChatService {
	return &ChatService{llm: client}
}

// Ask validates the question, builds a prompt, and generates an answer.
func (s *ChatService) Ask(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", ErrQuestionRequired
	}

	return s.llm.Generate(ctx, prompt.Build(question))
}
