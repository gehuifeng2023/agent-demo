package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agent-demo/internal/document"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/model"
	"agent-demo/internal/retriever"
)

func TestKnowledgeHandlerListReturnsAllChunks(t *testing.T) {
	unifiedRetriever := testKnowledgeRetriever()
	handler := NewKnowledgeHandler(unifiedRetriever)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp model.KnowledgeListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(resp.Chunks))
	}
	if resp.Chunks[0].Type != "knowledge_base" || resp.Chunks[0].KnowledgeBaseID != "default" {
		t.Fatalf("expected default knowledge chunk, got %#v", resp.Chunks[0])
	}
	if resp.Chunks[1].Type != "file" || resp.Chunks[1].FileID != "file-1" {
		t.Fatalf("expected uploaded file chunk, got %#v", resp.Chunks[1])
	}
}

func TestKnowledgeHandlerRecallReturnsQuestionMatches(t *testing.T) {
	unifiedRetriever := testKnowledgeRetriever()
	handler := NewKnowledgeHandler(unifiedRetriever)
	body := strings.NewReader(`{"question":"BetaProject 支持什么？","file_ids":["file-1"],"top_k":5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/retrieve", body)
	rr := httptest.NewRecorder()

	handler.Recall(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp model.KnowledgeRecallResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Question != "BetaProject 支持什么？" {
		t.Fatalf("expected trimmed question, got %q", resp.Question)
	}
	if len(resp.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(resp.Chunks))
	}
	if resp.Chunks[0].FileID != "file-1" {
		t.Fatalf("expected selected uploaded file, got %#v", resp.Chunks[0])
	}
}

func TestKnowledgeHandlerRecallUsesConfiguredDefaultTopK(t *testing.T) {
	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "GammaProject first", Position: 1},
		{ID: "file-2", Source: "uploads/file.txt", Content: "GammaProject second", Position: 2},
	})
	handler := NewKnowledgeHandlerWithTopK(unifiedRetriever, 1)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/retrieve", strings.NewReader(`{"question":"GammaProject"}`))
	rr := httptest.NewRecorder()

	handler.Recall(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp model.KnowledgeRecallResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Chunks) != 1 {
		t.Fatalf("expected 1 chunk from configured topK, got %d", len(resp.Chunks))
	}
}

func TestKnowledgeHandlerRecallRequiresQuestion(t *testing.T) {
	handler := NewKnowledgeHandler(testKnowledgeRetriever())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/retrieve", strings.NewReader(`{"question":" "}`))
	rr := httptest.NewRecorder()

	handler.Recall(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "question is empty") {
		t.Fatalf("expected question error, got body=%s", rr.Body.String())
	}
}

func TestKnowledgeHandlerRejectsWrongMethod(t *testing.T) {
	handler := NewKnowledgeHandler(testKnowledgeRetriever())

	listReq := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge", nil)
	listRR := httptest.NewRecorder()
	handler.List(listRR, listReq)
	if listRR.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected list status 405, got %d", listRR.Code)
	}

	recallReq := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/retrieve", nil)
	recallRR := httptest.NewRecorder()
	handler.Recall(recallRR, recallReq)
	if recallRR.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected recall status 405, got %d", recallRR.Code)
	}
}

func testKnowledgeRetriever() *retriever.UnifiedRetriever {
	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID: "default",
		Chunks: []document.Chunk{
			{ID: "kb-1", Source: "docs/faq.md", Content: "AlphaProject 支持默认知识库检索。", Position: 1},
		},
	})
	unifiedRetriever.StoreFileChunks("file-1", []document.Chunk{
		{ID: "file-1", Source: "uploads/file.txt", Content: "BetaProject 支持上传文件检索。", Position: 1},
	})
	return unifiedRetriever
}
