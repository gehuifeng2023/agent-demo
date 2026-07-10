package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileReaderTool struct {
	RootDir string
}

func (t FileReaderTool) Name() string {
	return "file_reader"
}
func (t FileReaderTool) Description() string {
	return "读取指定安全目录内的文件内容"
}

func (t FileReaderTool) Execute(ctx context.Context, input string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(input) {
		return "", fmt.Errorf("path not allowed")
	}

	root := strings.TrimSpace(t.RootDir)
	if root == "" {
		return "", fmt.Errorf("root dir is empty")
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root dir: %w", err)
	}
	full, err := filepath.Abs(filepath.Join(rootAbs, filepath.Clean(input)))
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	rel, err := filepath.Rel(rootAbs, full)
	if err != nil {
		return "", fmt.Errorf("check path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path not allowed")
	}

	info, err := os.Stat(full)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}

	data, err := os.ReadFile(full)
	return string(data), err
}
