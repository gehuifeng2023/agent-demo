package handler

import (
	"agent-demo/internal/prompt"
	"encoding/json"
	"net/http"

	"agent-demo/internal/agent"
	"agent-demo/internal/model"
)

type ChatHandler struct {
	agent *agent.Agent
}

func NewChatHandler(agent *agent.Agent) *ChatHandler {
	return &ChatHandler{
		agent: agent,
	}
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req model.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	promptType := prompt.TypeChat
	if req.Type != "" {
		promptType = prompt.Type(req.Type)
	}

	answer, answerType, err := h.agent.ChatWithType(r.Context(), req.Question, promptType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	resp := model.ChatResponse{
		Answer: answer,
		Type:   answerType,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
