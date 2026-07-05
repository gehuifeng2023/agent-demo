package session

import (
	"context"
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Store interface {
	Append(ctx context.Context, sessionID string, messages ...Message) error
	Recent(ctx context.Context, sessionID string, limit int) ([]Message, error)
}
