package session

import (
	"context"
	"time"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error
	GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error)
	GetSessionField(ctx context.Context, sessionID string, field string) (interface{}, error)
	UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
}
