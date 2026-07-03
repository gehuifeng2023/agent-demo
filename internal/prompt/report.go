package prompt

type reportTemplate struct{}

func (t reportTemplate) Build(req Request) Prompt {
	return Prompt{
		System:   reportSystemPrompt,
		Question: "请根据以下内容生成报告：\n" + req.Input,
	}
}

const reportSystemPrompt = `你是一个专业的报告生成助手。

请根据用户提供的信息生成中文报告。

回答要求：
1. 明确报告主题和结论
2. 按背景、进展、问题、建议组织
3. 事实和判断分开表达
4. 输出适合正式场景阅读`
