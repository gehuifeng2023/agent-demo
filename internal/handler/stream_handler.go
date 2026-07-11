package handler

import (
	"agent-demo/internal/agent"
	"agent-demo/internal/model"
	"encoding/json"
	"fmt"
	"net/http"
)

type StreamHandler struct {
	agent *agent.Agent
}

func NewStreamHandler(agent *agent.Agent) *StreamHandler {
	return &StreamHandler{agent: agent}
}

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req model.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.agent.Stream(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	metadata, err := json.Marshal(map[string]any{
		"sessionID": result.SessionID,
		"type":      result.Type,
		"sources":   result.Sources,
	})
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}
	fmt.Fprintf(w, "event: meta\ndata: %s\n\n", metadata)
	flusher.Flush()

	failed := false
	chunks, errs := result.Chunks, result.Errors
	for chunks != nil || errs != nil {
		select {
		case s, ok := <-chunks:
			if !ok {
				chunks = nil
				continue
			}
			payload, marshalErr := json.Marshal(s)
			if marshalErr != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", marshalErr.Error())
				failed = true
				flusher.Flush()
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		case err, ok := <-errs:
			if ok && err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				failed = true
				flusher.Flush()
			}
			errs = nil
		case <-r.Context().Done():
			return
		}
	}
	if !failed {
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}
}
