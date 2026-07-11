package embedding

import "context"

type Client interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}
