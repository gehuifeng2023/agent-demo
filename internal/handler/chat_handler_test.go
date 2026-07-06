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
	"agent-demo/internal/llm"
	"agent-demo/internal/model"
)

func TestChatHandlerReturnsSources(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	agentCore, err := agent.NewAgent(llm.NewMockClient())
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

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
	if resp.Sources[0].File != "docs/faq.md" {
		t.Fatalf("expected docs/faq.md, got %q", resp.Sources[0].File)
	}
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
