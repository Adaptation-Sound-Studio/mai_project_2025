package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"upload-service/internal/config"
)

func SetupTestRedis(t *testing.T) *redis.Client {
	cfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "Rbkkth3920",
	}
	client := NewRedisClient(cfg)

	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Redis должен быть запущен локально на 6379")

	err = client.FlushAll(ctx).Err()
	require.NoError(t, err, "Не удалось очистить Redis перед тестами")

	return client
}

func TestSetAndGetArtistID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := SetupTestRedis(t)

	ctx := context.Background()
	userID := "testuser"
	artistID := int64(42)

	initialSession := map[string]interface{}{
		"user_id": 123,
	}
	raw, _ := json.Marshal(initialSession)
	err := rdb.Set(ctx, userID, raw, 0).Err()
	require.NoError(t, err)

	err = SetArtistID(rdb, userID, strconv.FormatInt(artistID, 10))
	require.NoError(t, err)

	got, err := GetArtistID(rdb, userID)
	require.NoError(t, err)
	assert.Equal(t, "42", got)
}

func TestGetUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := SetupTestRedis(t)

	sessionID := "admintoken"

	session := map[string]interface{}{
		"user_id": 44,
	}
	data, err := json.Marshal(session)
	require.NoError(t, err)

	err = rdb.Set(ctx, sessionID, data, 0).Err()
	require.NoError(t, err)

	userID, err := GetUserID(rdb, ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, int64(44), userID)
}

func TestGetUserRole_Found(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	rdb := SetupTestRedis(t)

	sessionJSON := `{"role":"artist"}`
	err := rdb.Set(ctx, "55", sessionJSON, 0).Err()
	require.NoError(t, err)

	role, err := GetUserRole(rdb, "55")
	require.NoError(t, err)
	assert.Equal(t, "artist", role)
}

func TestGetUserRole_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := SetupTestRedis(t)

	_, err := GetUserRole(rdb, "unknown")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}
