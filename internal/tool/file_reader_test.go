package tool

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReaderToolReadsFileInsideRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "faq.md"), []byte("hello tool"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := FileReaderTool{RootDir: root}.Execute(context.Background(), "faq.md")
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}
	if got != "hello tool" {
		t.Fatalf("expected file content, got %q", got)
	}
}

func TestFileReaderToolRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Dir(root)
	if err := os.WriteFile(filepath.Join(parent, "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	_, err := FileReaderTool{RootDir: root}.Execute(context.Background(), "../secret.txt")
	if err == nil {
		t.Fatal("expected traversal to fail")
	}
	if !strings.Contains(err.Error(), "path not allowed") {
		t.Fatalf("expected path not allowed error, got %v", err)
	}
}

func TestFileReaderToolRejectsAbsolutePath(t *testing.T) {
	root := t.TempDir()

	_, err := FileReaderTool{RootDir: root}.Execute(context.Background(), filepath.Join(root, "faq.md"))
	if err == nil {
		t.Fatal("expected absolute path to fail")
	}
	if !strings.Contains(err.Error(), "path not allowed") {
		t.Fatalf("expected path not allowed error, got %v", err)
	}
}

func TestFileReaderToolRejectsDirectory(t *testing.T) {
	root := t.TempDir()

	_, err := FileReaderTool{RootDir: root}.Execute(context.Background(), ".")
	if err == nil {
		t.Fatal("expected directory to fail")
	}
	if !strings.Contains(err.Error(), "path is a directory") {
		t.Fatalf("expected directory error, got %v", err)
	}
}
