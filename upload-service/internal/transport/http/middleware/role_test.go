package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"upload-service/internal/config"
	appredis "upload-service/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func SetupTestRedis(t *testing.T) *redis.Client {
	err := godotenv.Load("C:/Users/User/Desktop/Project/mai_project_2025/.env")
	ctx := context.Background()
	cfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: os.Getenv("REDIS_PASSWORD"),
	}
	client := appredis.NewRedisClient(cfg)

	err = client.Ping(ctx).Err()
	require.NoError(t, err, "Redis должен быть запущен локально на 6379")

	err = client.FlushAll(ctx).Err()
	require.NoError(t, err, "Не удалось очистить Redis перед тестами")

	return client
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
	require.Contains(t, rr.Body.String(), "Session ID not found in cookies")
}

func TestRequireAnyRole_InvalidSession(t *testing.T) {
	rdb := SetupTestRedis(t)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "invalid-token")
	rr := httptest.NewRecorder()

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен быть вызван")
	}))

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Contains(t, rr.Body.String(), "Session ID not found in cookies")
}

func TestRequireAnyRole_MissingRoleInRedis(t *testing.T) {
	rdb := SetupTestRedis(t)
	ctx := context.Background()

	err := rdb.Set(ctx, "session:token123", `{"user_id":123}`, 0).Err()
	require.NoError(t, err)

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен вызываться при отсутствии роли")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "token123"})
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	require.Contains(t, rr.Body.String(), "Не удалось определить роль пользователя")
}

func TestRequireAnyRole_MissingCookie(t *testing.T) {
	rdb := SetupTestRedis(t)

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен вызываться без cookie")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Contains(t, rr.Body.String(), "Session ID not found")
}

func TestRequireAnyRole_InvalidRole(t *testing.T) {
	rdb := SetupTestRedis(t)
	ctx := context.Background()

	err := rdb.Set(ctx, "session:token123", `{"user_id":123}`, 0).Err()
	require.NoError(t, err)
	err = rdb.Set(ctx, "role:123", "user", 0).Err()
	require.NoError(t, err)

	protected := RequireAnyRole(rdb, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler не должен вызываться при неправильной роли")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "token123"})
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	require.Contains(t, rr.Body.String(), "Не удалось определить роль пользователя")
}
