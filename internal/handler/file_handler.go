package handler

import (
	"net/http"
	"os"

	"agent-demo/internal/converter"
	"agent-demo/internal/document"
	"agent-demo/internal/model"
	"agent-demo/internal/upload"
)

type fileChunkStore interface {
	StoreFileChunks(fileID string, chunks []document.Chunk)
}

type FileHandler struct {
	uploadService *upload.Service
	fileStore     fileChunkStore
	converters    *converter.Registry
}

func NewFileHandler(dir string, maxSize int64, fileStore fileChunkStore) *FileHandler {
	return &FileHandler{
		uploadService: upload.NewService(dir, maxSize),
		fileStore:     fileStore,
		converters:    converter.NewRegistry(converter.DOCXConverter{}, converter.TXTConverter{}, converter.MarkdownConverter{}),
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

	content, err := h.converters.Convert(r.Context(), path)
	if err != nil {
		_ = os.Remove(path)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	chunks := document.SplitByParagraph([]document.Document{
		{
			Source:  path,
			Content: content,
		},
	})

	if h.fileStore != nil {
		h.fileStore.StoreFileChunks(id, chunks)
	}

	writeJSON(w, http.StatusOK, model.UploadResponse{FileID: id, FileName: header.Filename, Size: header.Size})
}
