package model

type ChatRequest struct {
	Question         string   `json:"question"`
	Type             string   `json:"type,omitempty"`
	SessionID        string   `json:"sessionID,omitempty"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
	FileIDs          []string `json:"file_ids,omitempty"`
}

type ChatResponse struct {
	SessionID string   `json:"sessionID"`
	Answer    string   `json:"answer"`
	Type      string   `json:"type"`
	Sources   []Source `json:"sources,omitempty"`
}
