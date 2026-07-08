package converter

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type Registry struct{ converters []Converter }

func NewRegistry(items ...Converter) *Registry { return &Registry{converters: items} }

func (r *Registry) Convert(ctx context.Context, path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, c := range r.converters {
		if c.Support(ext) {
			log.Println("upload file Type: ", ext)
			return c.Convert(ctx, path)
		}
	}
	return "", fmt.Errorf("no converter for %s", ext)
}
