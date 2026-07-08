package retriever

import (
	"agent-demo/internal/document"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/upload"
)

type UnifiedRetriever struct {
	knowledge *knowledge.Registry
	files     *upload.FileDocumentStore
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
	var chunks []document.Chunk

	kbs := r.knowledge.GetMany(kbIDs)
	explicitSelection := len(kbIDs) > 0 || len(fileIDs) > 0
	if !explicitSelection {
		kbs = r.knowledge.All()
	}
	for _, kb := range kbs {
		chunks = append(chunks, kb.Chunks...)
	}

	if !explicitSelection {
		chunks = append(chunks, r.files.AllChunks()...)
	} else {
		for _, fid := range fileIDs {
			chunks = append(chunks, r.files.GetChunks(fid)...)
		}
	}

	return NewKeywordRetriever(chunks).Retrieve(q, topK)
}
