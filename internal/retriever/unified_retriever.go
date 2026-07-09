package retriever

import (
	"sort"
	"strconv"

	"agent-demo/internal/document"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/upload"
)

type UnifiedRetriever struct {
	knowledge *knowledge.Registry
	files     *upload.FileDocumentStore
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
	return &UnifiedRetriever{
		knowledge: knowledge.NewRegistry(),
		files:     upload.NewFileDocumentStore(),
	}
}

func (r *UnifiedRetriever) RegisterKnowledgeBase(kb *knowledge.KnowledgeBase) {
	r.knowledge.Register(kb)
}

func (r *UnifiedRetriever) StoreFileChunks(fileID string, chunks []document.Chunk) {
	r.files.Store(fileID, chunks)
}

func (r *UnifiedRetriever) Retrieve(q string, kbIDs, fileIDs []string, topK int) []document.Chunk {
	sourceChunks := r.selectedSourceChunks(kbIDs, fileIDs)
	chunks := make([]document.Chunk, 0, len(sourceChunks))
	for _, sourceChunk := range sourceChunks {
		chunks = append(chunks, sourceChunk.Chunk)
	}

	return NewKeywordRetriever(chunks).Retrieve(q, topK)
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
