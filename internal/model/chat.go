package model

type ChatRequest struct {
	Question string `json:"question"`
	Type     string `json:"type"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
	Type   string `json:"type"`
}
