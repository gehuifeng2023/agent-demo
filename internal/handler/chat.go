package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"agent-demo/internal/service"
)

type chatService interface {
	Ask(ctx context.Context, question string) (string, error)
}

// ChatHandler serves the chat HTTP API.
type ChatHandler struct {
	service chatService
}

// NewChatHandler creates a handler with its service dependency.
func NewChatHandler(service chatService) *ChatHandler {
	return &ChatHandler{service: service}
}

type chatRequest struct {
	Question string `json:"question"`
}

type chatResponse struct {
	Answer string `json:"answer"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ServeHTTP implements http.Handler.
func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	var req chatRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	answer, err := h.service.Ask(r.Context(), req.Question)
	if err != nil {
		if errors.Is(err, service.ErrQuestionRequired) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "question is required"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, chatResponse{Answer: answer})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
