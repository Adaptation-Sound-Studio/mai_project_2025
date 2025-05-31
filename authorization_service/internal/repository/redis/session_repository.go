package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth_service/internal/domain/session"

	"github.com/go-redis/redis/v8"
)

type SessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) session.SessionRepository {
	return &SessionRepository{client: client}
}

func (r *SessionRepository) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, sessionID, jsonData, expiration).Err()
}

func (r *SessionRepository) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	val, err := r.client.Get(ctx, sessionID).Result()
	if err != nil {
		return err
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return err
	}

	sessionData[field] = value

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, sessionID, jsonData, 0).Err()
}

func (r *SessionRepository) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
	val, err := r.client.Get(ctx, sessionID).Result()
	if err != nil {
		return 0, err
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return 0, err
	}

	rawID, ok := sessionData["user_id"]
	if !ok {
		return 0, fmt.Errorf("user_id не найден в сессии")
	}

	switch id := rawID.(type) {
	case float64:
		return int64(id), nil
	case int64:
		return id, nil
	default:
		return 0, fmt.Errorf("user_id имеет неподдерживаемый тип: %T", id)
	}
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, sessionID).Err()
}
