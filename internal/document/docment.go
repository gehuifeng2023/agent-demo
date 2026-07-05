package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Source  string
	Content string
}

func LoadFromDir(dir string) ([]Document, error) {
	var docs []Document

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		docs = append(docs, Document{
			Source:  path,
			Content: string(content),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk dir %s: %w", dir, err)
	}

	return docs, nil
}
