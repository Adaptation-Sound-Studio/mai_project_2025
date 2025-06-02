package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"upload-service/internal/config"
	appredis "upload-service/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupTestRedis(t *testing.T) *redis.Client {
	ctx := context.Background()
	cfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "Rbkkth3920",
	}
	client := appredis.NewRedisClient(cfg)

	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Redis должен быть запущен локально на 6379")

	err = client.FlushAll(ctx).Err()
	require.NoError(t, err, "Не удалось очистить Redis перед тестами")

	return client
}

func TestGetCurrentArtistID_NoAuthHeader(t *testing.T) {
	rdb := SetupTestRedis(t)
	req := httptest.NewRequest("GET", "/", nil)

	artistID, err := GetCurrentArtistID(req, rdb)

	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Equal(t, int64(0), artistID)
}

func TestGetCurrentArtistID_ArtistNotFound(t *testing.T) {
	ctx := context.Background()
	rdb := SetupTestRedis(t)

	// только сессия, без artist:user-1
	err := rdb.Set(ctx, "session:token-xyz", "user-1", 0).Err()
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "token-xyz")

	artistID, err := GetCurrentArtistID(req, rdb)

	assert.ErrorIs(t, err, ErrForbidden)
	assert.Equal(t, int64(0), artistID)
}

func TestGetCurrentArtistID_Success(t *testing.T) {
	ctx := context.Background()
	rdb := SetupTestRedis(t)

	err := rdb.Set(ctx, "session:token-123", "user-1", 0).Err()
	require.NoError(t, err)
	err = rdb.Set(ctx, "artist:user-1", "99", 0).Err()
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "token-123")

	artistID, err := GetCurrentArtistID(req, rdb)

	require.NoError(t, err)
	assert.Equal(t, int64(99), artistID)
}

func TestGetUserID_Success(t *testing.T) {
	ctx := context.Background()
	rdb := SetupTestRedis(t)

	err := rdb.Set(ctx, "session:token-1", "123", 0).Err()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "token-1")

	userID, userIDStr, err := GetUserID(req, rdb)
	require.NoError(t, err)
	assert.Equal(t, int64(123), userID)
	assert.Equal(t, "123", userIDStr)
}

func TestGetUserID_InvalidUserIDFormat(t *testing.T) {
	ctx := context.Background()
	rdb := SetupTestRedis(t)

	err := rdb.Set(ctx, "session:token-bad", "not-an-int", 0).Err()
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "token-bad")

	userID, userIDStr, err := GetUserID(req, rdb)

	assert.ErrorIs(t, err, ErrInternal)
	assert.Equal(t, int64(0), userID)
	assert.Equal(t, "", userIDStr)
}

func TestGetUserID_SessionNotFound(t *testing.T) {
	rdb := SetupTestRedis(t)

	// не устанавливаем ключ session:missing-token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "missing-token")

	userID, userIDStr, err := GetUserID(req, rdb)

	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Equal(t, int64(0), userID)
	assert.Equal(t, "", userIDStr)
}
