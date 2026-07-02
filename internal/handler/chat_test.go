package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-demo/internal/service"
)

type fakeChatService struct {
	answer string
	err    error
}

func (f fakeChatService) Ask(ctx context.Context, question string) (string, error) {
	return f.answer, f.err
}

func TestChatHandlerSuccess(t *testing.T) {
	h := NewChatHandler(fakeChatService{answer: "APISIX 是一个云原生 API 网关。"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{"question":"什么是 APISIX？"}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got["answer"] == "" {
		t.Fatalf("answer is empty")
	}
}

func TestChatHandlerInvalidJSON(t *testing.T) {
	h := NewChatHandler(fakeChatService{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assertError(t, rec, http.StatusBadRequest, "invalid request body")
}

func TestChatHandlerQuestionRequired(t *testing.T) {
	h := NewChatHandler(fakeChatService{err: service.ErrQuestionRequired})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{"question":" "}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assertError(t, rec, http.StatusBadRequest, "question is required")
}

func TestChatHandlerMethodNotAllowed(t *testing.T) {
	h := NewChatHandler(fakeChatService{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assertError(t, rec, http.StatusMethodNotAllowed, "method not allowed")
}

func TestChatHandlerInternalError(t *testing.T) {
	h := NewChatHandler(fakeChatService{err: errors.New("boom")})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{"question":"hello"}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assertError(t, rec, http.StatusInternalServerError, "internal server error")
}

func assertError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got["error"] != wantError {
		t.Fatalf("error = %q, want %q", got["error"], wantError)
	}
}
