package handler

import (
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

	answer, answerType, err := h.agent.Chat(r.Context(), req.Question, req.Type)
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
