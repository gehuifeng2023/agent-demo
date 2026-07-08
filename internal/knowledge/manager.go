package knowledge

import "agent-demo/internal/document"

type KnowledgeBase struct {
	ID     string
	Chunks []document.Chunk
}
