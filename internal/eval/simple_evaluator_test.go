package eval

import (
	"testing"

	"agent-demo/internal/model"
)

func TestSimpleEvaluatorReportsSourcesAndFullAnswerScore(t *testing.T) {
	quality := SimpleEvaluator{}.Evaluate("这是一个足够长的回答，用于验证带来源时的质量评分。", []model.Source{{File: "docs/faq.md"}})

	if !quality.HasSources {
		t.Fatal("expected sources to be reported")
	}
	if quality.Score != 0.8 {
		t.Fatalf("expected score 0.8, got %v", quality.Score)
	}
	if len(quality.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", quality.Warnings)
	}
}

func TestSimpleEvaluatorWarnsForMissingSourcesAndShortAnswer(t *testing.T) {
	quality := SimpleEvaluator{}.Evaluate("太短", nil)

	if quality.HasSources {
		t.Fatal("expected no sources")
	}
	if quality.Score != 0.3 {
		t.Fatalf("expected score 0.3, got %v", quality.Score)
	}
	if len(quality.Warnings) != 2 {
		t.Fatalf("expected two warnings, got %#v", quality.Warnings)
	}
}

func TestSimpleEvaluatorReportsUncertaintyRisk(t *testing.T) {
	quality := SimpleEvaluator{}.Evaluate("由于资料不足，无法确定该问题的最终答案，请补充更多上下文。", []model.Source{{File: "docs/faq.md"}})

	if quality.Score != 0.9 {
		t.Fatalf("expected score 0.9, got %v", quality.Score)
	}
	if len(quality.Warnings) != 1 || quality.Warnings[0] != "回答包含不确定性说明" {
		t.Fatalf("unexpected warnings: %#v", quality.Warnings)
	}
}
