package upload

import (
	"testing"

	"agent-demo/internal/document"
)

func TestFileDocumentStoreStoresChunksByFileID(t *testing.T) {
	store := NewFileDocumentStore()
	chunks := []document.Chunk{
		{ID: "file-1-1", Source: "uploads/file-1.txt", Content: "Alpha content", Position: 1},
	}

	store.Store("file-1", chunks)

	got := store.GetChunks("file-1")
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(got))
	}
	if got[0].ID != "file-1-1" {
		t.Fatalf("expected chunk file-1-1, got %q", got[0].ID)
	}

	got[0].Content = "changed"
	again := store.GetChunks("file-1")
	if again[0].Content != "Alpha content" {
		t.Fatalf("expected stored chunk copy to be unchanged, got %q", again[0].Content)
	}
}
