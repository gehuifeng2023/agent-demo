package document

import (
	"fmt"
	"strings"
)

type Chunk struct {
	ID       string
	Source   string
	Content  string
	Position int
}

func SplitByParagraph(docs []Document) []Chunk {
	var chunks []Chunk

	for _, doc := range docs {
		parts := strings.Split(doc.Content, "\n\n")

		for index, part := range parts {
			text := strings.TrimSpace(part)
			if text == "" {
				continue
			}

			chunks = append(chunks, Chunk{
				ID:       fmt.Sprintf("%s-%d", doc.Source, index),
				Source:   doc.Source,
				Content:  text,
				Position: index,
			})
		}
	}

	return chunks
}
