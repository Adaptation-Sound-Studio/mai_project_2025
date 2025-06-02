package redis

import (
	"auth_service/internal/domain/session"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func setupTestRedis() session.SessionRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Rbkkth3920",
	})
	return NewSessionRepository(client)
}

func TestRedisSessionManager_CreateAndGet(t *testing.T) {
	manager := setupTestRedis()
	ctx := context.Background()
	sessionID := "test-session-1"
	data := map[string]interface{}{
		"user_id": int64(42),
	}

	err := manager.CreateSession(ctx, sessionID, data, time.Minute)
	require.NoError(t, err)

	id, err := manager.GetUserIDFromSession(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, int64(42), id)
}

func TestRedisSessionManager_UpdateSessionField(t *testing.T) {
	manager := setupTestRedis()
	ctx := context.Background()
	sessionID := "test-session-2"
	data := map[string]interface{}{
		"user_id": int64(77),
	}

	_ = manager.CreateSession(ctx, sessionID, data, time.Minute)

	err := manager.UpdateSessionField(ctx, sessionID, "user_id", int64(88))
	require.NoError(t, err)

	id, err := manager.GetUserIDFromSession(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, int64(88), id)
}

func TestRedisSessionManager_DeleteSession(t *testing.T) {
	manager := setupTestRedis()
	ctx := context.Background()
	sessionID := "test-session-3"
	data := map[string]interface{}{
		"user_id": int64(99),
	}

	_ = manager.CreateSession(ctx, sessionID, data, time.Minute)

	err := manager.DeleteSession(ctx, sessionID)
	require.NoError(t, err)

	_, err = manager.GetUserIDFromSession(ctx, sessionID)
	require.Error(t, err)
}

func TestCreateSession_InvalidData(t *testing.T) {
	manager := setupTestRedis()

	data := map[string]interface{}{
		"invalid": make(chan int), // каналы не сериализуются
	}

	err := manager.CreateSession(context.Background(), "bad_json", data, time.Minute)
	require.Error(t, err)
}

func TestUpdateSessionField_SessionNotFound(t *testing.T) {
	manager := setupTestRedis()

	err := manager.UpdateSessionField(context.Background(), "nonexistent", "field", "value")
	require.Error(t, err)
}

func TestGetSessionField_Success(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "Rbkkth3920"})
	repo := NewSessionRepository(rdb)

	ctx := context.Background()
	sessionID := "test-session"
	data := map[string]interface{}{
		"user_id": float64(42),
		"role":    "admin",
	}
	jsonData, _ := json.Marshal(data)
	err := rdb.Set(ctx, sessionID, jsonData, time.Minute).Err()
	require.NoError(t, err)

	val, err := repo.GetSessionField(ctx, sessionID, "user_id")
	require.NoError(t, err)
	require.Equal(t, float64(42), val)
}

func TestGetSessionField_FieldNotFound(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "Rbkkth3920"})
	repo := NewSessionRepository(rdb)

	ctx := context.Background()
	sessionID := "test-session-missing-field"
	data := map[string]interface{}{
		"only_this": "value",
	}
	jsonData, _ := json.Marshal(data)
	err := rdb.Set(ctx, sessionID, jsonData, time.Minute).Err()
	require.NoError(t, err)

	val, err := repo.GetSessionField(ctx, sessionID, "missing")
	require.Nil(t, val)
	require.Error(t, err)
	require.Contains(t, err.Error(), `field "missing" not found`)
}

func TestGetSessionField_InvalidJSON(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "Rbkkth3920"})
	repo := NewSessionRepository(rdb)

	ctx := context.Background()
	sessionID := "test-session-invalid-json"
	err := rdb.Set(ctx, sessionID, "not-a-json", time.Minute).Err()
	require.NoError(t, err)

	val, err := repo.GetSessionField(ctx, sessionID, "some")
	require.Nil(t, val)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid character")
}

func TestGetSessionField_RedisError(t *testing.T) {
	// Используем клиент на несуществующем порту, чтобы сымитировать ошибку
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6390", Password: ""})
	repo := NewSessionRepository(rdb)

	ctx := context.Background()
	val, err := repo.GetSessionField(ctx, "any", "field")
	require.Nil(t, val)
	require.Error(t, err)
}

func TestDeleteSessionsByUserID_Success(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "Rbkkth3920", DB: 0})
	repo := NewSessionRepository(rdb)

	userID := int64(101)
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	sessionIDs := []string{"sess1", "sess2"}
	for _, sid := range sessionIDs {
		err := rdb.Set(ctx, "session:"+sid, `{"user_id":101}`, time.Minute).Err()
		require.NoError(t, err)
	}
	err := rdb.SAdd(ctx, userSessionsKey, sessionIDs).Err()
	require.NoError(t, err)

	err = repo.DeleteSessionsByUserID(ctx, userID)
	require.NoError(t, err)

	for _, sid := range sessionIDs {
		exists, _ := rdb.Exists(ctx, "session:"+sid).Result()
		require.Equal(t, int64(0), exists)
	}
	exists, _ := rdb.Exists(ctx, userSessionsKey).Result()
	require.Equal(t, int64(0), exists)
}

func TestDeleteSessionsByUserID_NoSessions(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "Rbkkth3920", DB: 0})
	repo := NewSessionRepository(rdb)

	userID := int64(202)
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	rdb.Del(ctx, userSessionsKey)

	err := repo.DeleteSessionsByUserID(ctx, userID)
	require.NoError(t, err)
}
