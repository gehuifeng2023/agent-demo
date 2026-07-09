package converter

import (
	"context"
	"github.com/nguyenthenguyen/docx"
)

type DOCXConverter struct{}

func (c DOCXConverter) Support(ext string) bool { return ext == ".docx" || ext == ".doc" }
func (c DOCXConverter) Convert(ctx context.Context, path string) (string, error) {
	doc, err := docx.ReadDocxFile(path)
	if err != nil {
		return "", err
	}
	defer doc.Close()
	return "# Word 文档\n\n" + doc.Editable().GetContent(), nil
}
