package prompt

import (
	"fmt"
	"strings"

	"agent-demo/internal/document"
)

func BuildRAGPrompt(question string, chunks []document.Chunk) string {
	var contextBuilder strings.Builder

	for i, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("【文档片段 %d】\n", i+1))
		contextBuilder.WriteString(fmt.Sprintf("来源：%s\n", chunk.Source))
		contextBuilder.WriteString(chunk.Content)
		contextBuilder.WriteString("\n\n")
	}

	return fmt.Sprintf(`你是一个基于知识库回答问题的问答助手。

请严格根据下面提供的文档片段回答用户问题。

如果文档片段中没有相关信息，请明确回答：
“根据当前文档，暂时无法确定。”

不要编造文档中没有的信息。

文档片段：
%s

用户问题：
%s

请输出清晰、准确、结构化的回答。`, contextBuilder.String(), question)
}
