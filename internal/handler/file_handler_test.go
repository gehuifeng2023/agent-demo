package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-demo/internal/agent"
	"agent-demo/internal/document"
	"agent-demo/internal/llm"
	"agent-demo/internal/model"
	"agent-demo/internal/retriever"
)

type failingIndexedFileStore struct{}

func (failingIndexedFileStore) StoreFileChunks(string, []document.Chunk) {}

func (failingIndexedFileStore) IndexFileChunks(context.Context, string, []document.Chunk) error {
	return errors.New("embedding unavailable")
}

func TestFileHandlerUploadAddsFileToChatRetrieval(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	unifiedRetriever := retriever.NewUnifiedRetriever()
	agentCore := agent.NewAgent(llm.NewMockClient(), unifiedRetriever)

	handler := NewFileHandler(t.TempDir(), 20<<20, unifiedRetriever)
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

func TestFileHandlerUploadCanBeSelectedByFileID(t *testing.T) {
	restore := chdirRepoRoot(t)
	defer restore()

	unifiedRetriever := retriever.NewUnifiedRetriever()
	agentCore := agent.NewAgent(llm.NewMockClient(), unifiedRetriever)

	handler := NewFileHandler(t.TempDir(), 20<<20, unifiedRetriever)
	body, contentType := multipartBody(t, "file", "notes.txt", "BetaProject 支持按文件选择检索。")
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

	chatHandler := NewChatHandler(agentCore)
	chatBody := fmt.Sprintf(`{"question":"BetaProject 支持什么？","file_ids":[%q]}`, uploadResp.FileID)
	chatReq := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(chatBody))
	chatRR := httptest.NewRecorder()

	chatHandler.ServeHTTP(chatRR, chatReq)

	if chatRR.Code != http.StatusOK {
		t.Fatalf("expected chat status 200, got %d body=%s", chatRR.Code, chatRR.Body.String())
	}

	var chatResp model.ChatResponse
	if err := json.Unmarshal(chatRR.Body.Bytes(), &chatResp); err != nil {
		t.Fatalf("decode chat response: %v", err)
	}
	if len(chatResp.Sources) != 1 {
		t.Fatalf("expected one uploaded file source, got body=%s", chatRR.Body.String())
	}
	if !strings.Contains(chatResp.Sources[0].File, uploadResp.FileID) {
		t.Fatalf("expected source to include uploaded file id %q, got %q", uploadResp.FileID, chatResp.Sources[0].File)
	}
}

func TestFileHandlerUploadRejectsUnsupportedFileType(t *testing.T) {
	handler := NewFileHandler(t.TempDir(), 20<<20, nil)
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

func TestFileHandlerUploadDocUsesWordConverterAndCleansFailedFile(t *testing.T) {
	uploadDir := t.TempDir()
	handler := NewFileHandler(uploadDir, 20<<20, nil)
	body, contentType := multipartBody(t, "file", "notes.doc", "not a word document")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	handler.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "unsupported file type") {
		t.Fatalf("expected .doc to use word converter, got body=%s", rr.Body.String())
	}

	var savedFiles []string
	if err := filepath.WalkDir(uploadDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			savedFiles = append(savedFiles, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk upload dir: %v", err)
	}
	if len(savedFiles) != 0 {
		t.Fatalf("expected failed conversion file to be removed, got %v", savedFiles)
	}
}

func TestFileHandlerUploadCleansFileWhenEmbeddingFails(t *testing.T) {
	uploadDir := t.TempDir()
	handler := NewFileHandler(uploadDir, 20<<20, failingIndexedFileStore{})
	body, contentType := multipartBody(t, "file", "notes.txt", "content to embed")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()

	handler.Upload(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d body=%s", rr.Code, rr.Body.String())
	}
	var savedFiles []string
	if err := filepath.WalkDir(uploadDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			savedFiles = append(savedFiles, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk upload dir: %v", err)
	}
	if len(savedFiles) != 0 {
		t.Fatalf("expected failed embedding file to be removed, got %v", savedFiles)
	}
}

func TestFileHandlerUploadRequiresFile(t *testing.T) {
	handler := NewFileHandler(t.TempDir(), 20<<20, nil)
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
