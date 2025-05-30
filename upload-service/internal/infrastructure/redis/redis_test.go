package redis

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"upload-service/internal/config"
)

func SetupTestRedis(t *testing.T) *redis.Client {
	cfg := &config.RedisConfig{
		Addr: "localhost:6379",
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

	userID := "testuser"
	artistID := int64(42)

	err := SetArtistID(rdb, userID, artistID)
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
	err := rdb.Set(ctx, "session:"+sessionID, "44", 0).Err()
	require.NoError(t, err)

	userID, err := GetUserID(rdb, sessionID)
	require.NoError(t, err)
	assert.Equal(t, "44", userID)
}

func TestGetUserRole_Found(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := SetupTestRedis(t)

	err := rdb.Set(ctx, "role:55", "artist", 0).Err()
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
	assert.Contains(t, err.Error(), "роль не найдена")
}
