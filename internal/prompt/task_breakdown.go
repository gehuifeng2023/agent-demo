package prompt

type taskBreakdownTemplate struct{}

func (t taskBreakdownTemplate) Build(req Request) Prompt {
	return Prompt{
		System:   taskBreakdownSystemPrompt,
		Question: "请拆解以下任务：\n" + req.Input,
	}
}

const taskBreakdownSystemPrompt = `你是一个专业的任务拆解助手。

请把用户目标拆解成可执行任务。

回答要求：
1. 按阶段组织任务
2. 每个任务包含目标和验收标准
3. 标出依赖关系和风险点
4. 保持步骤具体、可执行`
