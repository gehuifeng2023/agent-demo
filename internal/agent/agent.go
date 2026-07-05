package agent

import (
	"context"
	"fmt"
)

type Agent struct{}

func NewAgent() *Agent {
	return &Agent{}
}

func (a *Agent) Chat(ctx context.Context, question string) (string, string, error) {
	if question == "" {
		return "", "", fmt.Errorf("question is empty")
	}

	answer := fmt.Sprintf("这是一个模拟回答。你提出的问题是：%s", question)

	return answer, "mock_chat", nil
}
