package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"agent-demo/internal/model"
	"agent-demo/internal/retriever"
)

const defaultKnowledgeRecallTopK = 3

type KnowledgeHandler struct {
	retriever *retriever.UnifiedRetriever
}

func NewKnowledgeHandler(unifiedRetriever *retriever.UnifiedRetriever) *KnowledgeHandler {
	if unifiedRetriever == nil {
		unifiedRetriever = retriever.NewUnifiedRetriever()
	}
	return &KnowledgeHandler{retriever: unifiedRetriever}
}

func (h *KnowledgeHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, model.KnowledgeListResponse{
		Chunks: buildKnowledgeChunks(h.retriever.AllChunks()),
	})
}

func (h *KnowledgeHandler) Recall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req model.KnowledgeRecallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	question := strings.TrimSpace(req.Question)
	if question == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "question is empty"})
		return
	}

	topK := req.TopK
	if topK <= 0 {
		topK = defaultKnowledgeRecallTopK
	}

	chunks := h.retriever.RetrieveChunks(
		question,
		compactKnowledgeStrings(req.KnowledgeBaseIDs),
		compactKnowledgeStrings(req.FileIDs),
		topK,
	)
	writeJSON(w, http.StatusOK, model.KnowledgeRecallResponse{
		Question: question,
		Chunks:   buildKnowledgeChunks(chunks),
	})
}

func buildKnowledgeChunks(chunks []retriever.SourceChunk) []model.KnowledgeChunk {
	result := make([]model.KnowledgeChunk, 0, len(chunks))
	for _, sourceChunk := range chunks {
		result = append(result, model.KnowledgeChunk{
			Type:            string(sourceChunk.Type),
			KnowledgeBaseID: sourceChunk.KnowledgeBaseID,
			FileID:          sourceChunk.FileID,
			File:            sourceChunk.Chunk.Source,
			ChunkID:         sourceChunk.Chunk.ID,
			Content:         sourceChunk.Chunk.Content,
			Position:        sourceChunk.Chunk.Position,
		})
	}
	return result
}

func compactKnowledgeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
