package converter

import (
	"context"
	"os"
)

type MarkdownConverter struct{}

func (c MarkdownConverter) Support(ext string) bool { return ext == ".md" }
func (c MarkdownConverter) Convert(ctx context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

type TXTConverter struct{}

func (c TXTConverter) Support(ext string) bool { return ext == ".txt" }
func (c TXTConverter) Convert(ctx context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return "# 上传文本\n\n" + string(data), nil
}
