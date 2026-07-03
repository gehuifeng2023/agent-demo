package prompt

type summarizeTemplate struct{}

func (t summarizeTemplate) Build(req Request) Prompt {
	return Prompt{
		System:   summarizeSystemPrompt,
		Question: "请总结以下内容：\n" + req.Input,
	}
}

const summarizeSystemPrompt = `你是一个专业的摘要生成助手。

请用中文总结用户提供的内容。

回答要求：
1. 保留关键事实和结论
2. 删除重复和无关信息
3. 用条理清晰的结构输出
4. 不改变原文含义`
