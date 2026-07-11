package retriever

import (
	"context"
	"testing"

	"agent-demo/internal/document"
)

func TestVectorRetrieverRanksByCosineSimilarity(t *testing.T) {
	embedder := testEmbeddingClient{vectors: map[string][]float64{"query": {1, 0}}}
	retriever := NewVectorRetriever(embedder, []document.Chunk{
		{ID: "near"},
		{ID: "far"},
	}, [][]float64{{1, 0}, {0, 1}})

	chunks, err := retriever.Retrieve(context.Background(), "query", 1)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if len(chunks) != 1 || chunks[0].ID != "near" {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}
}
