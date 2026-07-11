package retriever

import (
	"agent-demo/internal/document"
	"context"
)

type Retriever interface {
	Retrieve(ctx context.Context, query string, topK int) ([]document.Chunk, error)
}
