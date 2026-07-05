package model

type ChatRequest struct {
	Question string `json:"question"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
	Type   string `json:"type"`
}
