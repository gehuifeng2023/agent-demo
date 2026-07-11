package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agent-demo/internal/agent"
	"agent-demo/internal/llm"
)

func TestStreamHandlerReturnsSSEEvents(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	handler := NewStreamHandler(agent.NewAgent(llm.NewMockClient(), defaultTestRetriever(t)))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", strings.NewReader(`{"question":"什么是 RAG？"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); got != "text/event-stream; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: meta\n") {
		t.Fatalf("expected metadata event, got %q", body)
	}
	if !strings.Contains(body, `data: "正在分析..."`) || !strings.Contains(body, `data: "这是流式输出示例。"`) {
		t.Fatalf("expected streamed answer, got %q", body)
	}
	if !strings.Contains(body, "data: [DONE]\n\n") {
		t.Fatalf("expected done event, got %q", body)
	}
}

func TestStreamHandlerRejectsInvalidJSON(t *testing.T) {
	handler := NewStreamHandler(agent.NewAgent(llm.NewMockClient(), nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}
