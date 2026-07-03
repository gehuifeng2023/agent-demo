package prompt

type chatTemplate struct{}

func (t chatTemplate) Build(req Request) Prompt {
	return Prompt{
		System:   chatSystemPrompt,
		Question: req.Input,
	}
}

type ragChatTemplate struct{}

func (t ragChatTemplate) Build(req Request) Prompt {
	return Prompt{
		System:   ragChatSystemPrompt,
		Question: "参考资料：\n" + req.Context + "\n\n问题：\n" + req.Input,
	}
}

const chatSystemPrompt = `你是一个专业的技术问答助手。

请用中文回答用户问题。

回答要求：
1. 先给结论
2. 再解释原因
3. 最后给出建议
4. 如果信息不足，请明确说明不确定点
5. 不要编造不存在的事实`

const ragChatSystemPrompt = `你是一个专业的基于资料回答的问答助手。

请只基于用户提供的参考资料回答问题。

回答要求：
1. 先给结论
2. 引用参考资料中的关键信息
3. 资料不足时明确说明无法从资料判断
4. 不要编造资料中不存在的事实`
