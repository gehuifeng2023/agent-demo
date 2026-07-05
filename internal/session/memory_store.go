package session

import (
	"context"
	"strings"
	"sync"
)

type MemoryStore struct {
	mu          sync.RWMutex
	messages    map[string][]Message
	maxMessages int
}

func NewMemoryStore(maxMessages int) *MemoryStore {
	if maxMessages <= 0 {
		maxMessages = 20
	}

	return &MemoryStore{
		messages:    make(map[string][]Message),
		maxMessages: maxMessages,
	}
}

func (s *MemoryStore) Append(ctx context.Context, sessionID string, messages ...Message) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || len(messages) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.messages[sessionID]
	existing = append(existing, messages...)

	if len(existing) > s.maxMessages {
		existing = existing[len(existing)-s.maxMessages:]
	}

	s.messages[sessionID] = existing

	return nil
}

func (s *MemoryStore) Recent(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	existing := s.messages[sessionID]
	if len(existing) == 0 {
		return nil, nil
	}

	if limit <= 0 || limit > len(existing) {
		limit = len(existing)
	}

	start := len(existing) - limit
	result := make([]Message, 0, limit)
	result = append(result, existing[start:]...)

	return result, nil
}
