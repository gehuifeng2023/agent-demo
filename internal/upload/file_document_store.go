package upload

import (
	"sync"

	"agent-demo/internal/document"
)

type FileDocumentStore struct {
	mu     sync.RWMutex
	chunks map[string][]document.Chunk
}

func NewFileDocumentStore() *FileDocumentStore {
	return &FileDocumentStore{
		chunks: map[string][]document.Chunk{},
	}
}

func (f *FileDocumentStore) Store(fileID string, chunks []document.Chunk) {
	if fileID == "" || len(chunks) == 0 {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.chunks[fileID] = append([]document.Chunk(nil), chunks...)
}

func (f *FileDocumentStore) GetChunks(fileID string) []document.Chunk {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return append([]document.Chunk(nil), f.chunks[fileID]...)
}

func (f *FileDocumentStore) AllChunks() []document.Chunk {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var out []document.Chunk
	for _, chunks := range f.chunks {
		out = append(out, chunks...)
	}
	return out
}

func (f *FileDocumentStore) AllByFileID() map[string][]document.Chunk {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := make(map[string][]document.Chunk, len(f.chunks))
	for fileID, chunks := range f.chunks {
		out[fileID] = append([]document.Chunk(nil), chunks...)
	}
	return out
}
