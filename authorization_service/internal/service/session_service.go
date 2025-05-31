package service

import (
	"auth_service/internal/domain/session"
	"context"
	"time"
)

type SessionService struct {
	repo session.SessionRepository
}

func NewSessionService(repo session.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	return s.repo.CreateSession(ctx, sessionID, data, expiration)
}

func (s *SessionService) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	return s.repo.UpdateSessionField(ctx, sessionID, field, value)
}

func (s *SessionService) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
	return s.repo.GetUserIDFromSession(ctx, sessionID)
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.repo.DeleteSession(ctx, sessionID)
}
