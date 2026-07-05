package prompt

import (
	"fmt"
	"strings"

	"agent-demo/internal/session"
)

func WithHistory(question string, history []session.Message) string {
	question = strings.TrimSpace(question)

	if len(history) == 0 {
		return question
	}

	var builder strings.Builder

	builder.WriteString("历史对话：\n")

	for i, msg := range history {
		role := "用户"
		if msg.Role == session.RoleAssistant {
			role = "助手"
		}

		builder.WriteString(fmt.Sprintf("%d. %s：%s\n", i+1, role, msg.Content))
	}

	builder.WriteString("\n当前问题：\n")
	builder.WriteString(question)

	return builder.String()
}
