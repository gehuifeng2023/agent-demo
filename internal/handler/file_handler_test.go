package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agent-demo/internal/agent"
	"agent-demo/internal/llm"
	"agent-demo/internal/model"
)

func TestFileHandlerUploadAddsFileToChatRetrieval(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	agentCore, err := agent.NewAgent(llm.NewMockClient())
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}

	handler := NewFileHandler(t.TempDir(), 20<<20, agentCore)
	body, contentType := multipartBody(t, "file", "notes.txt", "AlphaProject 支持文件上传检索。\n\n第二段内容。")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	handler.Upload(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var uploadResp model.UploadResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if uploadResp.FileID == "" {
		t.Fatalf("expected file_id, got body=%s", rr.Body.String())
	}

	chatHandler := NewChatHandler(agentCore)
	chatReq := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"question":"AlphaProject 支持什么？"}`))
	chatRR := httptest.NewRecorder()

	chatHandler.ServeHTTP(chatRR, chatReq)

	if chatRR.Code != http.StatusOK {
		t.Fatalf("expected chat status 200, got %d body=%s", chatRR.Code, chatRR.Body.String())
	}

	var chatResp model.ChatResponse
	if err := json.Unmarshal(chatRR.Body.Bytes(), &chatResp); err != nil {
		t.Fatalf("decode chat response: %v", err)
	}
	if len(chatResp.Sources) == 0 {
		t.Fatalf("expected uploaded file source, got body=%s", chatRR.Body.String())
	}
	if !strings.Contains(chatResp.Sources[0].File, uploadResp.FileID) {
		t.Fatalf("expected source to include uploaded file id %q, got %q", uploadResp.FileID, chatResp.Sources[0].File)
	}
}

func TestFileHandlerUploadRejectsUnsupportedFileType(t *testing.T) {
	agentCore := &agent.Agent{}
	handler := NewFileHandler(t.TempDir(), 20<<20, agentCore)
	body, contentType := multipartBody(t, "file", "notes.pdf", "not supported")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	handler.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "unsupported file type") {
		t.Fatalf("expected unsupported file type error, got body=%s", rr.Body.String())
	}
}

func TestFileHandlerUploadRequiresFile(t *testing.T) {
	agentCore := &agent.Agent{}
	handler := NewFileHandler(t.TempDir(), 20<<20, agentCore)
	body, contentType := multipartBody(t, "other", "notes.txt", "missing file field")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	handler.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "file is required") {
		t.Fatalf("expected required file error, got body=%s", rr.Body.String())
	}
}

func multipartBody(t *testing.T, fieldName string, fileName string, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}
