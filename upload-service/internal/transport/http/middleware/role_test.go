package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"upload-service/internal/config"
	appredis "upload-service/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func SetupTestRedis(t *testing.T) *redis.Client {
	ctx := context.Background()
	cfg := &config.RedisConfig{
		Addr: "localhost:6379",
	}
	client := appredis.NewRedisClient(cfg)

	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Redis должен быть запущен локально на 6379")

	err = client.FlushAll(ctx).Err()
	require.NoError(t, err, "Не удалось очистить Redis перед тестами")

	return client
}

func TestRequireAnyRole_ValidAccess_WithRealRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()

	rdb := SetupTestRedis(t)

	err := rdb.Set(ctx, "session:test-token", "123", 0).Err()
	require.NoError(t, err)
	err = rdb.Set(ctx, "role:123", "admin", 0).Err()
	require.NoError(t, err)

	handlerCalled := false
	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "test-token")
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, handlerCalled)
}

func TestRequireAnyRole_MissingAuthorizationHeader(t *testing.T) {
	rdb := SetupTestRedis(t)

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен быть вызван")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Contains(t, rr.Body.String(), "Не указан токен")
}

func TestRequireAnyRole_InvalidSession(t *testing.T) {
	rdb := SetupTestRedis(t)

	// session не установлен
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "invalid-token")
	rr := httptest.NewRecorder()

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен быть вызван")
	}))

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Contains(t, rr.Body.String(), "Сессия недействительна")
}

func TestRequireAnyRole_MissingRoleInRedis(t *testing.T) {
	ctx := context.Background()
	rdb := SetupTestRedis(t)

	// Создаём сессию, но не устанавливаем роль
	err := rdb.Set(ctx, "session:some-token", "456", 0).Err()
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "some-token")
	rr := httptest.NewRecorder()

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен быть вызван")
	}))

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	require.Contains(t, rr.Body.String(), "Не удалось определить роль пользователя")
}
