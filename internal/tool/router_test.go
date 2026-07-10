package tool

import "testing"

func TestRouteToolDetectsFileReadQuestion(t *testing.T) {
	if got := RouteTool("帮我读取 faq.md 并总结"); got != "file_reader" {
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
