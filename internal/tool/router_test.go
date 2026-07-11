package tool

import "testing"

func TestRouteToolDetectsFileReadQuestion(t *testing.T) {
	if got := RouteTool("帮我读取 faq.md 并总结"); got != "file_reader" {
		t.Fatalf("expected file_reader, got %q", got)
	}
}

func TestRouteToolDetectsLogAnalysisQuestion(t *testing.T) {
	question := "帮我分析日志 request_id=abc status=502 upstream timeout"
	if got := RouteTool(question); got != "log_analyzer" {
		t.Fatalf("expected log_analyzer, got %q", got)
	}
}

func TestRouteToolPrefersFileReaderForFileQuestion(t *testing.T) {
	if got := RouteTool("请读取 error.md 并总结"); got != "file_reader" {
		t.Fatalf("expected file_reader, got %q", got)
	}
}

func TestRouteToolIgnoresNormalQuestion(t *testing.T) {
	if got := RouteTool("什么是 RAG？"); got != "" {
		t.Fatalf("expected no tool, got %q", got)
	}
}

func TestExtractFilePath(t *testing.T) {
	got := ExtractFilePath("请查看 knowledge_attachment/default/faq.md，然后总结")
	if got != "knowledge_attachment/default/faq.md" {
		t.Fatalf("unexpected path %q", got)
	}
}

func TestExtractFilePathWithoutSpaces(t *testing.T) {
	got := ExtractFilePath("请读取faq.md并总结")
	if got != "faq.md" {
		t.Fatalf("unexpected path %q", got)
	}
}

func TestExtractToolInputForLogAnalyzer(t *testing.T) {
	question := "帮我分析日志 request_id=abc status=502"
	got := ExtractToolInput("log_analyzer", question)
	if got != question {
		t.Fatalf("unexpected input %q", got)
	}
}
