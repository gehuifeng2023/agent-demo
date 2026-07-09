package model

type KnowledgeChunk struct {
	Type            string `json:"type"`
	KnowledgeBaseID string `json:"knowledge_base_id,omitempty"`
	FileID          string `json:"file_id,omitempty"`
	File            string `json:"file"`
	ChunkID         string `json:"chunk_id"`
	Content         string `json:"content"`
	Position        int    `json:"position"`
}

type KnowledgeListResponse struct {
	Chunks []KnowledgeChunk `json:"chunks"`
}

type KnowledgeRecallRequest struct {
	Question         string   `json:"question"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
	FileIDs          []string `json:"file_ids,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
}

type KnowledgeRecallResponse struct {
	Question string           `json:"question"`
	Chunks   []KnowledgeChunk `json:"chunks"`
}
