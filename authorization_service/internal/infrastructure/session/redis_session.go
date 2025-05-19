package session

import (
	"context"
	"encoding/json"
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

func (r *RedisSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, sessionID).Err()
}
