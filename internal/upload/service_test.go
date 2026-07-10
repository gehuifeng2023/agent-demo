package upload

import (
	"io"
	"mime/multipart"
	"strings"
	"testing"
)

type readSeekCloser struct {
	*strings.Reader
}

func (r readSeekCloser) Close() error {
	return nil
}

func TestServiceSaveUsesDefaultMaxSizeWhenUnset(t *testing.T) {
	service := NewService(t.TempDir(), 0)
	file := readSeekCloser{Reader: strings.NewReader("hello")}
	header := &multipart.FileHeader{
		Filename: "notes.txt",
		Size:     5,
	}

	id, path, err := service.Save(file, header)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if id == "" || path == "" {
		t.Fatalf("expected saved file id and path, got id=%q path=%q", id, path)
	}
}

var _ multipart.File = readSeekCloser{Reader: strings.NewReader("")}
var _ io.Closer = readSeekCloser{Reader: strings.NewReader("")}
