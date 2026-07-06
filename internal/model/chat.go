package model

type ChatRequest struct {
	Question  string `json:"question"`
	Type      string `json:"type"`
	SessionID string `json:"sessionID"`
}

type ChatResponse struct {
	SessionID string   `json:"sessionID"`
	Answer    string   `json:"answer"`
	Type      string   `json:"type"`
	Sources   []Source `json:"sources,omitempty"`
}
