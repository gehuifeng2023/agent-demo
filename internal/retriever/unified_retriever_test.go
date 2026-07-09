package retriever

import (
	"testing"

	"agent-demo/internal/document"
	"agent-demo/internal/knowledge"
)

func TestUnifiedRetrieverUsesSelectedKnowledgeBase(t *testing.T) {
	unifiedRetriever := NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: []document.Chunk{{ID: "kb-1", Source: "docs/faq.md", Content: "RAG knowledge", Position: 1}},
	})
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "RAG file", Position: 1},
	})

	chunks := unifiedRetriever.Retrieve("RAG", []string{"default"}, nil, 3)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Source != "docs/faq.md" {
		t.Fatalf("expected selected knowledge source, got %q", chunks[0].Source)
	}
}

func TestUnifiedRetrieverUsesSelectedFile(t *testing.T) {
	unifiedRetriever := NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: []document.Chunk{{ID: "kb-1", Source: "docs/faq.md", Content: "Alpha knowledge", Position: 1}},
	})
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "Alpha file", Position: 1},
	})

	chunks := unifiedRetriever.Retrieve("Alpha", nil, []string{"file-1"}, 3)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Source != "uploads/file.txt" {
		t.Fatalf("expected selected file source, got %q", chunks[0].Source)
	}
}

func TestUnifiedRetrieverDefaultsToAllSources(t *testing.T) {
	unifiedRetriever := NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: []document.Chunk{{ID: "kb-1", Source: "docs/faq.md", Content: "Alpha knowledge", Position: 1}},
	})
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "Alpha file", Position: 1},
	})

	chunks := unifiedRetriever.Retrieve("Alpha", nil, nil, 3)

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
}
