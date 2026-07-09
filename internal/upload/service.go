package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Service struct {
	Dir     string
	MaxSize int64
}

func NewService(dir string, maxSize int64) *Service {
	return &Service{Dir: dir, MaxSize: maxSize}
}

func (s *Service) Save(file multipart.File, header *multipart.FileHeader) (id, path string, err error) {
	maxSize := s.MaxSize
	if maxSize <= 0 {
		maxSize = 20 << 20
	}
	if header.Size > s.MaxSize {
		return "", "", fmt.Errorf("file too large")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".md" && ext != ".txt" && ext != ".docx" && ext != ".doc" {
		return "", "", fmt.Errorf("unsupported file type: %s", ext)
	}

	id = fmt.Sprintf("f-%d", time.Now().UnixNano())
	dayDir := filepath.Join(s.Dir, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dayDir, 0755); err != nil {
		return "", "", err
	}
	path = filepath.Join(dayDir, id+ext)
	out, err := os.Create(path)
	if err != nil {
		return "", "", err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	return id, path, err
}
