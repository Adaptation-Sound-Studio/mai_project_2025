package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	err = r.client.Set(ctx, sessionID, jsonData, expiration).Err()
	if err != nil {
		return err
	}

	var userID int64
	switch v := data["user_id"].(type) {
	case int64:
		userID = v
	case float64:
		userID = int64(v)
	default:
		return fmt.Errorf("invalid or missing user_id in session data")
	}

	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)
	err = r.client.SAdd(ctx, userSessionsKey, sessionID).Err()
	if err != nil {
		return err
	}

	err = r.client.Expire(ctx, userSessionsKey, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepository) DeleteSessionsByUserID(ctx context.Context, userID int64) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		log.Printf("No sessions found for user %d", userID)
		return nil
	}

	log.Printf("Deleting sessions for user %d: %v", userID, sessionIDs)

	pipe := r.client.TxPipeline()
	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, "session:"+sessionID)
	}
	pipe.Del(ctx, userSessionsKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	log.Printf("Sessions deleted for user %d", userID)
	return nil
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

func (r *SessionRepository) GetSessionField(ctx context.Context, sessionID string, field string) (interface{}, error) {
	val, err := r.client.Get(ctx, sessionID).Result()
	if err != nil {
		return nil, err
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(val), &sessionData); err != nil {
		return nil, err
	}

	value, ok := sessionData[field]
	if !ok {
		return nil, fmt.Errorf("field %q not found in session", field)
	}

	return value, nil
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
