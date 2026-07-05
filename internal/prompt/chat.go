package prompt

import "fmt"

type ChatBuilder struct{}

func NewChatBuilder() *ChatBuilder {
	return &ChatBuilder{}
}

func (b *ChatBuilder) Build(input string) string {
	return fmt.Sprintf(`你是一个专业的技术问答智能体。

请用中文回答用户问题。

回答要求：
1. 先给结论
2. 再解释原因
3. 最后给出建议
4. 如果信息不足，请明确说明不确定点
5. 不要编造不存在的事实

用户问题：
%s`, input)
}
