package retriever

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"agent-demo/internal/document"
)

var nonAlnumRe = regexp.MustCompile(`[^\p{L}\p{N}]+`)

type KeywordRetriever struct {
	mu     sync.RWMutex
	chunks []document.Chunk
}

func NewKeywordRetriever(chunks []document.Chunk) *KeywordRetriever {
	return &KeywordRetriever{
		chunks: chunks,
	}
}
func (r *KeywordRetriever) Retrieve(query string, topK int) []document.Chunk {
	query = normalizeText(query)
	if query == "" || topK <= 0 {
		return nil
	}

	r.mu.RLock()
	chunks := append([]document.Chunk(nil), r.chunks...)
	r.mu.RUnlock()

	type scoredChunk struct {
		chunk document.Chunk
		score int
	}

	var scored []scoredChunk

	terms := strings.Fields(query)

	for _, chunk := range chunks {
		content := normalizeText(chunk.Content)

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

func (r *KeywordRetriever) AddChunks(chunks []document.Chunk) {
	if len(chunks) == 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.chunks = append(r.chunks, chunks...)
}

func normalizeText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	text = nonAlnumRe.ReplaceAllString(text, " ")
	return strings.Join(strings.Fields(text), " ")
}
