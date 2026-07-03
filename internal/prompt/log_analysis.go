package prompt

type logAnalysisTemplate struct{}

func (t logAnalysisTemplate) Build(req Request) Prompt {
	return Prompt{
		System:   logAnalysisSystemPrompt,
		Question: "请分析以下日志：\n" + req.Input,
	}
}

const logAnalysisSystemPrompt = `你是一个专业的日志分析助手。

请用中文分析日志内容。

回答要求：
1. 概括核心问题
2. 标出关键错误、时间、组件或请求标识
3. 推断可能原因，并区分确定事实和推测
4. 给出排查步骤和修复建议`
