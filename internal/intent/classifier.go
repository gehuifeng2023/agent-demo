package intent

import (
	"agent-demo/internal/prompt"
	"context"
	"strings"
)

type Classifier struct{}

func NewClassifier() *Classifier {
	return &Classifier{}
}

func (c *Classifier) Classify(ctx context.Context, input string) (prompt.Type, error) {
	text := strings.ToLower(strings.TrimSpace(input))

	if text == "" {
		return prompt.TypeChat, nil
	}

	if looksLikeLog(text) {
		return prompt.TypeLogAnalysis, nil
	}

	if looksLikeSummarizeTask(text) {
		return prompt.TypeSummarize, nil
	}

	if looksLikeTaskBreakdown(text) {
		return prompt.TypeTaskBreakdown, nil
	}

	return prompt.TypeChat, nil
}

func looksLikeLog(text string) bool {
	keywords := []string{
		"error",
		"warn",
		"panic",
		"exception",
		"timeout",
		"failed",
		"502",
		"500",
		"401",
		"403",
		"405",
		"signaturedoesnotmatch",
		"connection refused",
		"connection reset",
		"broken pipe",
		"no such file or directory",
		"stack trace",
		"request_id",
		"status",
		"duration_ms",
		"trace_id",
		"apisix",
		"nginx",
		"kubernetes",
		"pod",
		"evicted",
		"custom-forward-auth",
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func looksLikeSummarizeTask(text string) bool {
	keywords := []string{
		"总结",
		"摘要",
		"概括",
		"提炼",
		"归纳",
		"summarize",
		"summary",
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

func looksLikeTaskBreakdown(text string) bool {
	keywords := []string{
		"任务拆解",
		"拆解任务",
		"开发计划",
		"实施计划",
		"需求拆解",
		"里程碑",
		"排期",
		"todo",
		"task breakdown",
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}
