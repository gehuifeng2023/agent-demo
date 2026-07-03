package prompt

import (
	"strings"
	"testing"
)

func TestDefaultBuilderBuildsSupportedPromptTypes(t *testing.T) {
	tests := []struct {
		name       string
		request    Request
		wantSystem string
		wantUser   string
	}{
		{
			name:       "question answer",
			request:    Request{Type: TypeQA, Input: "  什么是 APISIX？  "},
			wantSystem: "技术问答助手",
			wantUser:   "什么是 APISIX？",
		},
		{
			name:       "rag question answer",
			request:    Request{Type: TypeRAGQA, Input: "如何配置限流？", Context: "APISIX 支持 limit-count 插件。"},
			wantSystem: "基于资料回答",
			wantUser:   "参考资料：\nAPISIX 支持 limit-count 插件。\n\n问题：\n如何配置限流？",
		},
		{
			name:       "log analysis",
			request:    Request{Type: TypeLogAnalysis, Input: "ERROR timeout"},
			wantSystem: "日志分析助手",
			wantUser:   "请分析以下日志：\nERROR timeout",
		},
		{
			name:       "summary",
			request:    Request{Type: TypeSummary, Input: "第一段\n第二段"},
			wantSystem: "摘要生成助手",
			wantUser:   "请总结以下内容：\n第一段\n第二段",
		},
		{
			name:       "task breakdown",
			request:    Request{Type: TypeTaskBreakdown, Input: "重构 prompt 模块"},
			wantSystem: "任务拆解助手",
			wantUser:   "请拆解以下任务：\n重构 prompt 模块",
		},
		{
			name:       "report",
			request:    Request{Type: TypeReport, Input: "本周完成 prompt 重构"},
			wantSystem: "报告生成助手",
			wantUser:   "请根据以下内容生成报告：\n本周完成 prompt 重构",
		},
	}

	builder := NewDefaultBuilder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builder.Build(tt.request)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if !strings.Contains(got.System, tt.wantSystem) {
				t.Fatalf("Build() system = %q, want containing %q", got.System, tt.wantSystem)
			}
			if got.Question != tt.wantUser {
				t.Fatalf("Build() question = %q, want %q", got.Question, tt.wantUser)
			}
		})
	}
}

func TestDefaultBuilderUsesQAWhenTypeIsEmpty(t *testing.T) {
	got, err := NewDefaultBuilder().Build(Request{Input: "  hello  "})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if got.Question != "hello" {
		t.Fatalf("Build() question = %q, want hello", got.Question)
	}
	if !strings.Contains(got.System, "技术问答助手") {
		t.Fatalf("Build() system = %q, want QA system prompt", got.System)
	}
}

func TestDefaultBuilderRejectsUnsupportedType(t *testing.T) {
	_, err := NewDefaultBuilder().Build(Request{Type: Type("unknown"), Input: "hello"})
	if err == nil {
		t.Fatal("Build() error = nil, want unsupported prompt type error")
	}
}

func TestBuildKeepsBackwardCompatibleQAHelper(t *testing.T) {
	got := Build("  hello  ")

	if got.Question != "hello" {
		t.Fatalf("Build() question = %q, want hello", got.Question)
	}
	if !strings.Contains(got.System, "技术问答助手") {
		t.Fatalf("Build() system = %q, want QA system prompt", got.System)
	}
}
