package retriever

import (
	"agent-demo/internal/document"
	"agent-demo/internal/embedding"
	"context"
	"fmt"
	"sort"
	"strings"
)

type VectorRetriever struct {
	embedder embedding.Client
	chunks   []document.Chunk
	vectors  [][]float64
}

func NewVectorRetriever(embedder embedding.Client, chunks []document.Chunk, vectors [][]float64) *VectorRetriever {
	return &VectorRetriever{embedder: embedder, chunks: chunks, vectors: vectors}
}

func (r *VectorRetriever) Retrieve(ctx context.Context, query string, topK int) ([]document.Chunk, error) {
	if r.embedder == nil {
		return nil, fmt.Errorf("embedding client is nil")
	}
	if strings.TrimSpace(query) == "" || topK <= 0 {
		return nil, nil
	}
	if len(r.chunks) != len(r.vectors) {
		return nil, fmt.Errorf("chunk count=%d does not match vector count=%d", len(r.chunks), len(r.vectors))
	}
	qv, err := r.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed retrieval query: %w", err)
	}
	if len(qv) != 1 || len(qv[0]) == 0 {
		return nil, fmt.Errorf("embed retrieval query: expected one non-empty vector")
	}
	type scored struct {
		idx   int
		score float64
	}
	list := make([]scored, 0, len(r.vectors))
	for i, v := range r.vectors {
		if len(v) != len(qv[0]) {
			return nil, fmt.Errorf("embedding dimension mismatch for chunk %q: got %d, want %d", r.chunks[i].ID, len(v), len(qv[0]))
		}
		list = append(list, scored{i, Cosine(qv[0], v)})
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score == list[j].score {
			return list[i].idx < list[j].idx
		}
		return list[i].score > list[j].score
	})
	if len(list) > topK {
		list = list[:topK]
	}
	out := make([]document.Chunk, 0, len(list))
	for _, x := range list {
		out = append(out, r.chunks[x.idx])
	}
	return out, nil
}
