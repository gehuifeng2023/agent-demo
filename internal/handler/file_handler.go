package handler

import (
	"agent-demo/internal/agent"
	"agent-demo/internal/document"
	"agent-demo/internal/model"
	"agent-demo/internal/upload"
	"net/http"
	"os"
)

type FileHandler struct {
	uploadService *upload.Service
	agent         *agent.Agent
}

func NewFileHandler(dir string, maxSize int64, agentCore *agent.Agent) *FileHandler {
	return &FileHandler{
		uploadService: upload.NewService(dir, maxSize),
		agent:         agentCore,
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	maxSize := h.uploadService.MaxSize
	if maxSize <= 0 {
		maxSize = 20 << 20
	}

	if err := r.ParseMultipartForm(maxSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	id, path, err := h.uploadService.Save(file, header)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	chunks := document.SplitByParagraph([]document.Document{
		{
			Source:  path,
			Content: string(content),
		},
	})

	if h.agent != nil {
		h.agent.AddDocumentChunks(chunks)
	}

	writeJSON(w, http.StatusOK, model.UploadResponse{FileID: id, FileName: header.Filename, Size: header.Size})
}
