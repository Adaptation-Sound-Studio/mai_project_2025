package session

import (
	"context"
	"time"
)

// SessionManager — интерфейс для работы с сессиями
type SessionManager interface {
	CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error
	UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
}
