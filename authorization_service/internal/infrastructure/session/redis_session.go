package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisSessionManager struct {
	client *redis.Client
}

func NewRedisSessionManager(client *redis.Client) *RedisSessionManager {
	return &RedisSessionManager{client: client}
}

func (r *RedisSessionManager) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, sessionID, jsonData, expiration).Err()
}

func (r *RedisSessionManager) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
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

func (r *RedisSessionManager) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
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

func (r *RedisSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, sessionID).Err()
}
