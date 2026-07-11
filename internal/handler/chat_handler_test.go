package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"agent-demo/internal/agent"
	"agent-demo/internal/document"
	"agent-demo/internal/knowledge"
	"agent-demo/internal/llm"
	"agent-demo/internal/model"
	"agent-demo/internal/retriever"
)

func TestChatHandlerReturnsSources(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	agentCore := agent.NewAgent(llm.NewMockClient(), defaultTestRetriever(t))

	handler := NewChatHandler(agentCore)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"question":"什么是 RAG？"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp model.ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Sources) == 0 {
		t.Fatalf("expected sources in response, got body=%s", rr.Body.String())
	}
	if resp.Sources[0].File != "knowledge_attachment/default/faq.md" {
		t.Fatalf("expected knowledge_attachment/default/faq.md, got %q", resp.Sources[0].File)
	}
	if !resp.Quality.HasSources {
		t.Fatalf("expected quality to report sources, got %#v", resp.Quality)
	}
	if resp.Quality.Score <= 0 {
		t.Fatalf("expected positive quality score, got %#v", resp.Quality)
	}
}

func TestChatHandlerUsesSelectedKnowledgeBase(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	agentCore := agent.NewAgent(llm.NewMockClient(), defaultTestRetriever(t))

	handler := NewChatHandler(agentCore)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"question":"什么是 RAG？","knowledge_base_ids":["default"]}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp model.ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Sources) == 0 {
		t.Fatalf("expected sources in response, got body=%s", rr.Body.String())
	}
	if resp.Sources[0].File != "knowledge_attachment/default/faq.md" {
		t.Fatalf("expected knowledge_attachment/default/faq.md, got %q", resp.Sources[0].File)
	}
}

func defaultTestRetriever(t *testing.T) *retriever.UnifiedRetriever {
	t.Helper()

	docs, err := document.LoadFromDir("knowledge_attachment/default/")
	if err != nil {
		t.Fatalf("load default knowledge: %v", err)
	}

	unifiedRetriever := retriever.NewUnifiedRetriever()
	unifiedRetriever.RegisterKnowledgeBase(&knowledge.KnowledgeBase{
		ID:     "default",
		Chunks: document.SplitByParagraph(docs),
	})
	return unifiedRetriever
}

func chdirRepoRoot(t *testing.T) func() {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir repo root: %v", err)
	}

	return func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}
