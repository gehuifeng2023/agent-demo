package agent

import (
	"agent-demo/internal/prompt"
	"context"
	"fmt"

	"agent-demo/internal/llm"
)

type Agent struct {
	llmClient     llm.Client
	promptFactory *prompt.Factory
}

func NewAgent(llmClient llm.Client) *Agent {
	return &Agent{
		llmClient:     llmClient,
		promptFactory: prompt.NewFactory(),
	}
}

func (a *Agent) Chat(ctx context.Context, question string) (string, string, error) {
	if question == "" {
		return "", "", fmt.Errorf("question is empty")
	}

	promptTxt := buildPrompt(question)
	answer, err := a.llmClient.Generate(ctx, promptTxt)
	if err != nil {
		return "", "", fmt.Errorf("generate answer: %w", err)
	}

	return answer, "llm_chat", nil
}

func buildPrompt(question string) string {
	return fmt.Sprintf(`
你是一个专业的技术问答智能体。
请用中文回答用户问题。
回答要求：
1. 先给结论
2. 再解释原因
3. 最后给出建议
4. 如果信息不足，请说明不确定点

用户问题：
%s`, question)
}

func (a *Agent) ChatWithType(ctx context.Context, question string, promptType prompt.Type) (string, string, error) {
	if question == "" {
		return "", "", fmt.Errorf("question is empty")
	}

	promptText, err := a.promptFactory.Build(promptType, question)
	if err != nil {
		return "", "", fmt.Errorf("build prompt: %w", err)
	}

	answer, err := a.llmClient.Generate(ctx, promptText)
	if err != nil {
		return "", "", fmt.Errorf("generate answer: %w", err)
	}

	return answer, string(promptType), nil
}
