package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiKeyMiddleware_ValidKey(t *testing.T) {
	expectedKey := "secret123"

	handlerCalled := false

	handler := ApiKeyMiddleware(expectedKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", expectedKey)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, handlerCalled)
}

func TestApiKeyMiddleware_InvalidKey(t *testing.T) {
	expectedKey := "secret123"

	handler := ApiKeyMiddleware(expectedKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called with invalid key")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", "wrongkey")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestApiKeyMiddleware_MissingKey(t *testing.T) {
	expectedKey := "secret123"

	handler := ApiKeyMiddleware(expectedKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when key is missing")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRoleFromContext_WithRole(t *testing.T) {
	ctx := context.WithValue(context.Background(), roleKey, "admin")
	role, ok := RoleFromContext(ctx)

	require.True(t, ok)
	require.Equal(t, "admin", role)
}

func TestRoleFromContext_WithoutRole(t *testing.T) {
	ctx := context.Background()
	role, ok := RoleFromContext(ctx)

	require.False(t, ok)
	require.Equal(t, "", role)
}

func TestHasRole_Allowed(t *testing.T) {
	ok := hasRole("moderator", []string{"user", "moderator", "admin"})
	require.True(t, ok)
}

func TestHasRole_NotAllowed(t *testing.T) {
	ok := hasRole("guest", []string{"user", "moderator", "admin"})
	require.False(t, ok)
}
