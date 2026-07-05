package prompt

import (
	"fmt"
	"strings"

	"agent-demo/internal/document"
	"agent-demo/internal/session"
)

func BuildRAGPrompt(question string, chunks []document.Chunk, history []session.Message) string {
	var docBuilder strings.Builder

	for i, chunk := range chunks {
		docBuilder.WriteString(fmt.Sprintf("【文档片段 %d】\n", i+1))
		docBuilder.WriteString(fmt.Sprintf("来源：%s\n", chunk.Source))
		docBuilder.WriteString(chunk.Content)
		docBuilder.WriteString("\n\n")
	}

	conversationText := WithHistory(question, history)

	return fmt.Sprintf(`你是一个基于知识库回答问题的助手。

请严格根据下面提供的文档片段和历史对话回答用户当前问题。

要求：
1. 优先依据文档片段回答
2. 可以结合历史对话理解代词、简称、上下文指代
3. 不要编造文档中没有的信息
4. 如果文档片段中没有答案，请回答“根据当前文档，暂时无法确定。”
5. 回答要清晰、准确、结构化
6. 尽量说明依据来自哪个文档

文档片段：
%s

对话上下文和当前问题：
%s`, docBuilder.String(), conversationText)
}
