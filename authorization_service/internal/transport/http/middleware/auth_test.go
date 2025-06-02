package middleware

import (
	"auth_service/internal/service"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type sessionRepoMock struct {
	GetUserIDFunc     func(ctx context.Context, sessionID string) (int64, error)
	GetSessionFieldFn func(ctx context.Context, sessionID string, field string) (interface{}, error)
}

func (m *sessionRepoMock) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	return nil // не используется в тестах
}

func (m *sessionRepoMock) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
	if m.GetUserIDFunc != nil {
		return m.GetUserIDFunc(ctx, sessionID)
	}
	return 0, errors.New("not implemented")
}

func (m *sessionRepoMock) GetSessionField(ctx context.Context, sessionID string, field string) (interface{}, error) {
	return nil, nil
}

func (m *sessionRepoMock) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	return nil
}

func (m *sessionRepoMock) DeleteSession(ctx context.Context, sessionID string) error {
	return nil
}

func (m *sessionRepoMock) DeleteSessionsByUserID(ctx context.Context, userID int64) error {
	return nil
}

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	repo := &sessionRepoMock{}
	svc := service.NewSessionService(repo)
	middleware := AuthMiddleware(svc)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_InvalidSession(t *testing.T) {
	repo := &sessionRepoMock{
		GetUserIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 0, errors.New("session not found")
		},
	}
	svc := service.NewSessionService(repo)
	middleware := AuthMiddleware(svc)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for invalid session")
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_ValidSession(t *testing.T) {
	const expectedUserID int64 = 42

	repo := &sessionRepoMock{
		GetUserIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
			if sessionID == "valid" {
				return expectedUserID, nil
			}
			return 0, errors.New("not found")
		},
	}
	svc := service.NewSessionService(repo)
	middleware := AuthMiddleware(svc)

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		uid, ok := UserIDFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, expectedUserID, uid)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
