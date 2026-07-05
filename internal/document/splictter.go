package document

import (
	"strings"
)

type Chunk struct {
	Source  string
	Content string
}

func SplitByParagraph(docs []Document) []Chunk {
	var chunks []Chunk

	for _, doc := range docs {
		parts := strings.Split(doc.Content, "\n\n")

		for _, part := range parts {
			text := strings.TrimSpace(part)
			if text == "" {
				continue
			}

			chunks = append(chunks, Chunk{
				Source:  doc.Source,
				Content: text,
			})
		}
	}

	return chunks
}
