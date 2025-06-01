package session

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func setupTestRedis() *RedisSessionManager {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return NewRedisSessionManager(client)
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
