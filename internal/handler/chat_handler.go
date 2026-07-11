package handler

import (
	"encoding/json"
	"net/http"

	"agent-demo/internal/agent"
	"agent-demo/internal/eval"
	"agent-demo/internal/model"
)

type ChatHandler struct {
	agent     *agent.Agent
	evaluator eval.Evaluator
}

func NewChatHandler(agent *agent.Agent) *ChatHandler {
	return NewChatHandlerWithEvaluator(agent, nil)
}

func NewChatHandlerWithEvaluator(agent *agent.Agent, evaluator eval.Evaluator) *ChatHandler {
	if evaluator == nil {
		evaluator = eval.SimpleEvaluator{}
	}
	return &ChatHandler{
		agent:     agent,
		evaluator: evaluator,
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

	answer, answerType, sessionID, sources, err := h.agent.Chat(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	resp := model.ChatResponse{
		SessionID: sessionID,
		Answer:    answer,
		Type:      answerType,
		Sources:   sources,
		Quality:   h.evaluator.Evaluate(answer, sources),
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
