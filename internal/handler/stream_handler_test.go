package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agent-demo/internal/agent"
	"agent-demo/internal/llm"
)

type failingStreamClient struct{}

func (failingStreamClient) Generate(context.Context, string) (string, error) {
	return "", errors.New("not used")
}

func (failingStreamClient) Stream(context.Context, string) (<-chan string, <-chan error) {
	chunks := make(chan string)
	close(chunks)
	errs := make(chan error, 1)
	errs <- errors.New("model stream failed")
	close(errs)
	return chunks, errs
}

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
	qualityIndex := strings.Index(body, "event: quality\n")
	doneIndex := strings.Index(body, "data: [DONE]\n\n")
	if qualityIndex == -1 || qualityIndex > doneIndex {
		t.Fatalf("expected quality event before done, got %q", body)
	}
	if !strings.Contains(body, `"has_sources":true`) {
		t.Fatalf("expected quality source status, got %q", body)
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

func TestStreamHandlerDoesNotEmitQualityForFailedStream(t *testing.T) {
	handler := NewStreamHandler(agent.NewAgent(failingStreamClient{}, nil))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", strings.NewReader(`{"question":"test"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "event: error\n") {
		t.Fatalf("expected error event, got %q", body)
	}
	if strings.Contains(body, "event: quality\n") || strings.Contains(body, "data: [DONE]\n\n") {
		t.Fatalf("did not expect quality or done after stream error, got %q", body)
	}
}
