package retriever

import (
	"testing"

	"agent-demo/internal/document"
)

func TestKeywordRetrieverRetrieveNormalizesPunctuation(t *testing.T) {
	retriever := NewKeywordRetriever([]document.Chunk{
		{
			ID:       "docs/faq.md-0",
			Source:   "docs/faq.md",
			Content:  "RAG 是 Retrieval-Augmented Generation 的缩写。",
			Position: 0,
		},
	})

	got := retriever.Retrieve("什么是 RAG？", 3)
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(got))
	}
	if got[0].Source != "docs/faq.md" {
		t.Fatalf("expected source docs/faq.md, got %q", got[0].Source)
	}
}
