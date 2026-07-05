package prompt

import "fmt"

type SummarizeBuilder struct{}

func NewSummarizeBuilder() *SummarizeBuilder {
	return &SummarizeBuilder{}
}

func (b *SummarizeBuilder) Build(input string) string {
	return fmt.Sprintf(`你是一个专业的技术文档摘要助手。

请对下面内容进行摘要。

摘要要求：
1. 先给一句话总结
2. 提取核心观点
3. 提取关键流程
4. 提取重要配置或参数
5. 提取风险点和注意事项
6. 最后给出适合研发人员阅读的简洁版本

内容：
%s`, input)
}
