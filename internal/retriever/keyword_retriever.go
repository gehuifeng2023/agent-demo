package retriever

import (
	"sort"
	"strings"

	"agent-demo/internal/document"
)

type KeywordRetriever struct {
	chunks []document.Chunk
}

func NewKeywordRetriever(chunks []document.Chunk) *KeywordRetriever {
	return &KeywordRetriever{
		chunks: chunks,
	}
}

func (r *KeywordRetriever) Retrieve(query string, topK int) []document.Chunk {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" || topK <= 0 {
		return nil
	}

	type scoredChunk struct {
		chunk document.Chunk
		score int
	}

	var scored []scoredChunk

	terms := strings.Fields(query)

	for _, chunk := range r.chunks {
		content := strings.ToLower(chunk.Content)

		score := 0
		for _, term := range terms {
			if strings.Contains(content, term) {
				score++
			}
		}

		if score > 0 {
			scored = append(scored, scoredChunk{
				chunk: chunk,
				score: score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if len(scored) > topK {
		scored = scored[:topK]
	}

	result := make([]document.Chunk, 0, len(scored))
	for _, item := range scored {
		result = append(result, item.chunk)
	}

	return result
}
