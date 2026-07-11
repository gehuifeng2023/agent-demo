package retriever

import (
	"context"
	"errors"
	"testing"

	"agent-demo/internal/document"
	"agent-demo/internal/knowledge"
)

type testEmbeddingClient struct {
	vectors map[string][]float64
	err     error
}

func (c testEmbeddingClient) Embed(_ context.Context, texts []string) ([][]float64, error) {
	if c.err != nil {
		return nil, c.err
	}
	result := make([][]float64, len(texts))
	for index, text := range texts {
		vector, ok := c.vectors[text]
		if !ok {
			return nil, errors.New("missing test vector")
		}
		result[index] = vector
	}
	return result, nil
}

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

func TestUnifiedRetrieverAllChunksIncludesSourceMetadata(t *testing.T) {
	unifiedRetriever := NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: []document.Chunk{{ID: "kb-1", Source: "docs/faq.md", Content: "Alpha knowledge", Position: 1}},
	})
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "Alpha file", Position: 1},
	})

	chunks := unifiedRetriever.AllChunks()

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Type != SourceTypeKnowledgeBase || chunks[0].KnowledgeBaseID != "default" {
		t.Fatalf("expected default knowledge chunk metadata, got %#v", chunks[0])
	}
	if chunks[1].Type != SourceTypeFile || chunks[1].FileID != "file-1" {
		t.Fatalf("expected uploaded file chunk metadata, got %#v", chunks[1])
	}
}

func TestUnifiedRetrieverRetrieveChunksPreservesSourceMetadata(t *testing.T) {
	unifiedRetriever := NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: []document.Chunk{{ID: "kb-1", Source: "docs/faq.md", Content: "Alpha knowledge", Position: 1}},
	})
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "Alpha file", Position: 1},
	})

	chunks := unifiedRetriever.RetrieveChunks("Alpha", nil, []string{"file-1"}, 3)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Type != SourceTypeFile || chunks[0].FileID != "file-1" {
		t.Fatalf("expected selected file chunk metadata, got %#v", chunks[0])
	}
	if chunks[0].Chunk.Source != "uploads/file.txt" {
		t.Fatalf("expected uploaded file source, got %q", chunks[0].Chunk.Source)
	}
}

func TestUnifiedRetrieverUsesVectorsWhenEmbeddingIsConfigured(t *testing.T) {
	embedder := testEmbeddingClient{vectors: map[string][]float64{
		"network database": {1, 0},
		"fruit recipe":     {0, 1},
		"find database":    {1, 0},
	}}
	unifiedRetriever := NewUnifiedRetrieverWithEmbedding(embedder)
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID: "default",
		Chunks: []document.Chunk{
			{ID: "network", Source: "docs/network.md", Content: "network database", Position: 1},
			{ID: "fruit", Source: "docs/fruit.md", Content: "fruit recipe", Position: 1},
		},
	})
	if err := unifiedRetriever.BuildVectorIndex(context.Background()); err != nil {
		t.Fatalf("build vector index: %v", err)
	}

	chunks, err := unifiedRetriever.RetrieveWithContext(context.Background(), "find database", nil, nil, 1)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if len(chunks) != 1 || chunks[0].ID != "network" {
		t.Fatalf("expected vector-ranked network chunk, got %#v", chunks)
	}
}

func TestUnifiedRetrieverReturnsEmbeddingError(t *testing.T) {
	embedder := testEmbeddingClient{err: errors.New("embedding unavailable")}
	unifiedRetriever := NewUnifiedRetrieverWithEmbedding(embedder)
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: []document.Chunk{{ID: "kb-1", Source: "docs/faq.md", Content: "content", Position: 1}},
	})
	if err := unifiedRetriever.BuildVectorIndex(context.Background()); err == nil {
		t.Fatal("expected index error")
	}
}
