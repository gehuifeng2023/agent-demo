package retriever

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"agent-demo/internal/document"
	"agent-demo/internal/embedding"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/upload"
)

type UnifiedRetriever struct {
	knowledge *knowledge.Registry
	files     *upload.FileDocumentStore
	embedder  embedding.Client
	vectorMu  sync.RWMutex
	vectors   map[string][]float64
}

type SourceType string

const (
	SourceTypeKnowledgeBase SourceType = "knowledge_base"
	SourceTypeFile          SourceType = "file"
)

type SourceChunk struct {
	Type            SourceType
	KnowledgeBaseID string
	FileID          string
	Chunk           document.Chunk
}

func NewUnifiedRetriever() *UnifiedRetriever {
	return NewUnifiedRetrieverWithEmbedding(nil)
}

func NewUnifiedRetrieverWithEmbedding(embedder embedding.Client) *UnifiedRetriever {
	return &UnifiedRetriever{
		knowledge: knowledge.NewRegistry(),
		files:     upload.NewFileDocumentStore(),
		embedder:  embedder,
		vectors:   make(map[string][]float64),
	}
}

func (r *UnifiedRetriever) RegisterKnowledgeBase(kb *knowledge.KnowledgeBase) {
	r.knowledge.Register(kb)
}

func (r *UnifiedRetriever) StoreFileChunks(fileID string, chunks []document.Chunk) {
	r.files.Store(fileID, chunks)
}

func (r *UnifiedRetriever) BuildVectorIndex(ctx context.Context) error {
	if r.embedder == nil {
		return nil
	}
	sourceChunks := r.selectedSourceChunks(nil, nil)
	vectors, err := embedSourceChunks(ctx, r.embedder, sourceChunks)
	if err != nil {
		return fmt.Errorf("embed knowledge chunks: %w", err)
	}

	r.vectorMu.Lock()
	r.vectors = vectors
	r.vectorMu.Unlock()
	return nil
}

func (r *UnifiedRetriever) IndexFileChunks(ctx context.Context, fileID string, chunks []document.Chunk) error {
	if r.embedder == nil {
		r.files.Store(fileID, chunks)
		return nil
	}

	sourceChunks := make([]SourceChunk, 0, len(chunks))
	for _, chunk := range chunks {
		sourceChunks = append(sourceChunks, SourceChunk{Type: SourceTypeFile, FileID: fileID, Chunk: chunk})
	}
	vectors, err := embedSourceChunks(ctx, r.embedder, sourceChunks)
	if err != nil {
		return fmt.Errorf("embed uploaded file chunks: %w", err)
	}

	r.files.Store(fileID, chunks)
	r.vectorMu.Lock()
	for key, vector := range vectors {
		r.vectors[key] = vector
	}
	r.vectorMu.Unlock()
	return nil
}

func (r *UnifiedRetriever) Retrieve(q string, kbIDs, fileIDs []string, topK int) []document.Chunk {
	sourceChunks := r.selectedSourceChunks(kbIDs, fileIDs)
	chunks := make([]document.Chunk, 0, len(sourceChunks))
	for _, sourceChunk := range sourceChunks {
		chunks = append(chunks, sourceChunk.Chunk)
	}

	return NewKeywordRetriever(chunks).Retrieve(q, topK)
}

func (r *UnifiedRetriever) RetrieveWithContext(ctx context.Context, q string, kbIDs, fileIDs []string, topK int) ([]document.Chunk, error) {
	sourceChunks, err := r.retrieveSourceChunks(ctx, q, kbIDs, fileIDs, topK)
	if err != nil {
		return nil, err
	}
	result := make([]document.Chunk, 0, len(sourceChunks))
	for _, sourceChunk := range sourceChunks {
		result = append(result, sourceChunk.Chunk)
	}
	return result, nil
}

func (r *UnifiedRetriever) AllChunks() []SourceChunk {
	return r.selectedSourceChunks(nil, nil)
}

func (r *UnifiedRetriever) RetrieveChunks(q string, kbIDs, fileIDs []string, topK int) []SourceChunk {
	sourceChunks := r.selectedSourceChunks(kbIDs, fileIDs)
	chunks := make([]document.Chunk, 0, len(sourceChunks))
	sourceByChunkID := make(map[string]SourceChunk, len(sourceChunks))
	for _, sourceChunk := range sourceChunks {
		chunks = append(chunks, sourceChunk.Chunk)
		sourceByChunkID[sourceChunkKey(sourceChunk.Chunk)] = sourceChunk
	}

	retrieved := NewKeywordRetriever(chunks).Retrieve(q, topK)
	result := make([]SourceChunk, 0, len(retrieved))
	for _, chunk := range retrieved {
		if sourceChunk, ok := sourceByChunkID[sourceChunkKey(chunk)]; ok {
			result = append(result, sourceChunk)
		}
	}
	return result
}

func (r *UnifiedRetriever) RetrieveChunksWithContext(ctx context.Context, q string, kbIDs, fileIDs []string, topK int) ([]SourceChunk, error) {
	return r.retrieveSourceChunks(ctx, q, kbIDs, fileIDs, topK)
}

func (r *UnifiedRetriever) retrieveSourceChunks(ctx context.Context, q string, kbIDs, fileIDs []string, topK int) ([]SourceChunk, error) {
	sourceChunks := r.selectedSourceChunks(kbIDs, fileIDs)
	if r.embedder == nil {
		return keywordSourceChunks(sourceChunks, q, topK), nil
	}
	if topK <= 0 || len(sourceChunks) == 0 {
		return nil, nil
	}

	queryVectors, err := r.embedder.Embed(ctx, []string{q})
	if err != nil {
		return nil, fmt.Errorf("embed retrieval query: %w", err)
	}
	if len(queryVectors) != 1 || len(queryVectors[0]) == 0 {
		return nil, fmt.Errorf("embed retrieval query: expected one non-empty vector")
	}

	r.vectorMu.RLock()
	defer r.vectorMu.RUnlock()
	type scoredChunk struct {
		sourceChunk SourceChunk
		score       float64
		index       int
	}
	scored := make([]scoredChunk, 0, len(sourceChunks))
	for index, sourceChunk := range sourceChunks {
		vector, ok := r.vectors[sourceVectorKey(sourceChunk)]
		if !ok {
			return nil, fmt.Errorf("missing embedding for chunk %q", sourceChunk.Chunk.ID)
		}
		if len(vector) != len(queryVectors[0]) {
			return nil, fmt.Errorf("embedding dimension mismatch for chunk %q: got %d, want %d", sourceChunk.Chunk.ID, len(vector), len(queryVectors[0]))
		}
		scored = append(scored, scoredChunk{sourceChunk: sourceChunk, score: Cosine(queryVectors[0], vector), index: index})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].index < scored[j].index
		}
		return scored[i].score > scored[j].score
	})
	if len(scored) > topK {
		scored = scored[:topK]
	}
	result := make([]SourceChunk, 0, len(scored))
	for _, item := range scored {
		result = append(result, item.sourceChunk)
	}
	return result, nil
}

func keywordSourceChunks(sourceChunks []SourceChunk, q string, topK int) []SourceChunk {
	chunks := make([]document.Chunk, 0, len(sourceChunks))
	sourceByChunkID := make(map[string]SourceChunk, len(sourceChunks))
	for _, sourceChunk := range sourceChunks {
		chunks = append(chunks, sourceChunk.Chunk)
		sourceByChunkID[sourceChunkKey(sourceChunk.Chunk)] = sourceChunk
	}

	retrieved := NewKeywordRetriever(chunks).Retrieve(q, topK)
	result := make([]SourceChunk, 0, len(retrieved))
	for _, chunk := range retrieved {
		if sourceChunk, ok := sourceByChunkID[sourceChunkKey(chunk)]; ok {
			result = append(result, sourceChunk)
		}
	}
	return result
}

func embedSourceChunks(ctx context.Context, embedder embedding.Client, sourceChunks []SourceChunk) (map[string][]float64, error) {
	if len(sourceChunks) == 0 {
		return map[string][]float64{}, nil
	}
	texts := make([]string, 0, len(sourceChunks))
	for _, sourceChunk := range sourceChunks {
		texts = append(texts, sourceChunk.Chunk.Content)
	}
	vectors, err := embedder.Embed(ctx, texts)
	if err != nil {
		return nil, err
	}
	if len(vectors) != len(sourceChunks) {
		return nil, fmt.Errorf("embedding count=%d, want %d", len(vectors), len(sourceChunks))
	}

	result := make(map[string][]float64, len(sourceChunks))
	dimension := 0
	for index, vector := range vectors {
		if len(vector) == 0 {
			return nil, fmt.Errorf("embedding for chunk %q is empty", sourceChunks[index].Chunk.ID)
		}
		if dimension == 0 {
			dimension = len(vector)
		} else if len(vector) != dimension {
			return nil, fmt.Errorf("embedding dimension=%d for chunk %q, want %d", len(vector), sourceChunks[index].Chunk.ID, dimension)
		}
		result[sourceVectorKey(sourceChunks[index])] = append([]float64(nil), vector...)
	}
	return result, nil
}

func (r *UnifiedRetriever) selectedSourceChunks(kbIDs, fileIDs []string) []SourceChunk {
	var chunks []SourceChunk

	kbs := r.knowledge.GetMany(kbIDs)
	explicitSelection := len(kbIDs) > 0 || len(fileIDs) > 0
	if !explicitSelection {
		kbs = r.knowledge.All()
	}
	for _, kb := range kbs {
		for _, chunk := range kb.Chunks {
			chunks = append(chunks, SourceChunk{
				Type:            SourceTypeKnowledgeBase,
				KnowledgeBaseID: kb.ID,
				Chunk:           chunk,
			})
		}
	}

	if !explicitSelection {
		for fileID, fileChunks := range r.files.AllByFileID() {
			for _, chunk := range fileChunks {
				chunks = append(chunks, SourceChunk{
					Type:   SourceTypeFile,
					FileID: fileID,
					Chunk:  chunk,
				})
			}
		}
	} else {
		for _, fid := range fileIDs {
			for _, chunk := range r.files.GetChunks(fid) {
				chunks = append(chunks, SourceChunk{
					Type:   SourceTypeFile,
					FileID: fid,
					Chunk:  chunk,
				})
			}
		}
	}

	sort.Slice(chunks, func(i, j int) bool {
		left := chunks[i]
		right := chunks[j]
		if sourceTypeRank(left.Type) != sourceTypeRank(right.Type) {
			return sourceTypeRank(left.Type) < sourceTypeRank(right.Type)
		}
		if left.KnowledgeBaseID != right.KnowledgeBaseID {
			return left.KnowledgeBaseID < right.KnowledgeBaseID
		}
		if left.FileID != right.FileID {
			return left.FileID < right.FileID
		}
		if left.Chunk.Source != right.Chunk.Source {
			return left.Chunk.Source < right.Chunk.Source
		}
		if left.Chunk.Position != right.Chunk.Position {
			return left.Chunk.Position < right.Chunk.Position
		}
		return left.Chunk.ID < right.Chunk.ID
	})

	return chunks
}

func sourceTypeRank(sourceType SourceType) int {
	switch sourceType {
	case SourceTypeKnowledgeBase:
		return 0
	case SourceTypeFile:
		return 1
	default:
		return 2
	}
}

func sourceChunkKey(chunk document.Chunk) string {
	return chunk.ID + "\x00" + chunk.Source + "\x00" + strconv.Itoa(chunk.Position)
}

func sourceVectorKey(sourceChunk SourceChunk) string {
	return string(sourceChunk.Type) + "\x00" + sourceChunk.KnowledgeBaseID + "\x00" + sourceChunk.FileID + "\x00" + sourceChunkKey(sourceChunk.Chunk)
}
