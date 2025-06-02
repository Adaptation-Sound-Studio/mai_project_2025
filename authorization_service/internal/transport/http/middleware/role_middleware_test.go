package middleware

import (
	"auth_service/internal/service"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleMiddleware_MissingCookie(t *testing.T) {
	repo := &sessionRepoMock{}
	svc := service.NewSessionService(repo)
	middleware := RoleMiddleware(svc, "admin")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRoleMiddleware_InvalidSession(t *testing.T) {
	repo := &sessionRepoMock{
		GetUserIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 0, errors.New("invalid session")
		},
	}
	svc := service.NewSessionService(repo)
	middleware := RoleMiddleware(svc, "admin")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRoleMiddleware_MissingRole(t *testing.T) {
	repo := &sessionRepoMock{
		GetUserIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 1, nil
		},
		GetSessionFieldFn: func(ctx context.Context, sessionID string, field string) (interface{}, error) {
			return nil, errors.New("role not set")
		},
	}
	svc := service.NewSessionService(repo)
	middleware := RoleMiddleware(svc, "admin")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess1"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRoleMiddleware_ForbiddenRole(t *testing.T) {
	repo := &sessionRepoMock{
		GetUserIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 1, nil
		},
		GetSessionFieldFn: func(ctx context.Context, sessionID string, field string) (interface{}, error) {
			return "user", nil
		},
	}
	svc := service.NewSessionService(repo)
	middleware := RoleMiddleware(svc, "admin")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess1"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
