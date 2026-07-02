package service

import (
	"context"
	"errors"
	"testing"

	"agent-demo/internal/prompt"
)

type fakeLLM struct {
	seen prompt.Prompt
}

func (f *fakeLLM) Generate(ctx context.Context, p prompt.Prompt) (string, error) {
	f.seen = p
	return "ok", nil
}

func TestChatServiceAskBuildsPromptAndReturnsAnswer(t *testing.T) {
	client := &fakeLLM{}
	svc := NewChatService(client)

	answer, err := svc.Ask(context.Background(), "  hello  ")
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if answer != "ok" {
		t.Fatalf("Ask() answer = %q, want ok", answer)
	}
	if client.seen.Question != "hello" {
		t.Fatalf("Ask() prompt question = %q, want hello", client.seen.Question)
	}
}

func TestChatServiceAskRequiresQuestion(t *testing.T) {
	svc := NewChatService(&fakeLLM{})

	_, err := svc.Ask(context.Background(), " ")
	if !errors.Is(err, ErrQuestionRequired) {
		t.Fatalf("Ask() error = %v, want ErrQuestionRequired", err)
	}
}
