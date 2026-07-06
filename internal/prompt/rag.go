package prompt

import (
	"fmt"
	"strings"

	"agent-demo/internal/document"
	"agent-demo/internal/session"
)

func BuildRAGPrompt(question string, chunks []document.Chunk, history []session.Message) string {
	documentText := buildDocumentText(chunks)
	conversationText := WithHistory(question, history)

	return fmt.Sprintf(`你是一个基于企业文档和知识库回答问题的助手。

你的任务是：严格依据【文档片段】回答【当前问题】。

【重要规则】
1. 事实性结论只能来自文档片段。
2. 历史对话只能用于理解上下文指代，不能作为事实依据。
3. 如果文档片段没有提供答案，必须回答：“根据当前文档，暂时无法确定。”
4. 不要使用模型自身知识补充业务事实。
5. 不要猜测、扩展或编造文档中没有的信息。
6. 如果多个文档片段互相矛盾，需要说明“当前文档信息不一致”，并分别列出依据。
7. 如果只能回答一部分，需要说明哪些内容可以确认，哪些内容无法确认。

【文档片段】
%s

【历史对话和当前问题】
%s

【输出格式】
请按以下结构输出：

1. 结论
   - 用 1 到 3 句话直接回答问题。
   - 如果无法确定，直接说明无法确定。

2. 依据
   - 列出支持结论的文档片段来源。
   - 格式：来源：xxx，依据：xxx。

3. 说明
   - 如果有不确定点，在这里说明。
   - 如果文档没有提到，也要明确指出。

请开始回答。`, documentText, conversationText)
}

func buildDocumentText(chunks []document.Chunk) string {
	if len(chunks) == 0 {
		return "未检索到相关文档片段。"
	}

	var builder strings.Builder

	for i, chunk := range chunks {
		builder.WriteString(fmt.Sprintf("【文档片段 %d】\n", i+1))
		builder.WriteString(fmt.Sprintf("来源：%s\n", chunk.Source))
		builder.WriteString("内容：\n")
		builder.WriteString(strings.TrimSpace(chunk.Content))
		builder.WriteString("\n\n")
	}

	return builder.String()
}
