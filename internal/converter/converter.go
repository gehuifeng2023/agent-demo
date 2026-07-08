package converter

import "context"

type Converter interface {
	Support(ext string) bool
	Convert(ctx context.Context, filePath string) (string, error)
}
