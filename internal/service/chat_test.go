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

type fakePromptBuilder struct {
	seen prompt.Request
}

func (f *fakePromptBuilder) Build(req prompt.Request) (prompt.Prompt, error) {
	f.seen = req
	return prompt.Prompt{
		System:   "system from builder",
		Question: "question from builder",
	}, nil
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

func TestChatServiceAskUsesPromptBuilder(t *testing.T) {
	client := &fakeLLM{}
	builder := &fakePromptBuilder{}
	svc := NewChatServiceWithPromptBuilder(client, builder)

	answer, err := svc.Ask(context.Background(), "  hello  ")
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if answer != "ok" {
		t.Fatalf("Ask() answer = %q, want ok", answer)
	}
	if builder.seen.Type != prompt.TypeQA {
		t.Fatalf("Ask() prompt type = %q, want %q", builder.seen.Type, prompt.TypeQA)
	}
	if builder.seen.Input != "  hello  " {
		t.Fatalf("Ask() builder input = %q, want original question", builder.seen.Input)
	}
	if client.seen.System != "system from builder" || client.seen.Question != "question from builder" {
		t.Fatalf("Ask() llm prompt = %#v, want prompt from builder", client.seen)
	}
}

func TestChatServiceAskRequiresQuestion(t *testing.T) {
	svc := NewChatService(&fakeLLM{})

	_, err := svc.Ask(context.Background(), " ")
	if !errors.Is(err, ErrQuestionRequired) {
		t.Fatalf("Ask() error = %v, want ErrQuestionRequired", err)
	}
}
