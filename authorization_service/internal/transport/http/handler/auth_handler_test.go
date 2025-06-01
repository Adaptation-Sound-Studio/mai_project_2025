package handler

import (
	"auth_service/internal/domain/user"
	"auth_service/internal/service"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser_Success(t *testing.T) {
	repo := &mockRepo{
		CreateFunc: func(u *user.User) error {
			require.Equal(t, "John", u.Name)
			return nil
		},
	}
	svc := service.NewAuthService(repo, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	body := []byte(`{"name":"John","login":"john","password":"1234"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestLogin_Success(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)

	repo := &mockRepo{
		GetByLoginFunc: func(login string) (*user.User, error) {
			return &user.User{
				ID:    1,
				Login: "john",
				Pass:  string(hashed),
			}, nil
		},
	}

	svc := service.NewAuthService(repo, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	body := []byte(`{"login":"john","password":"1234"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, resp.Header.Get("Set-Cookie"), "session_id=")
}

func TestLogin_InvalidBody(t *testing.T) {
	svc := service.NewAuthService(&mockRepo{}, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	body := []byte(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLogout_MissingCookie(t *testing.T) {
	svc := service.NewAuthService(&mockRepo{}, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogout_Success(t *testing.T) {
	var deletedSessionID string

	sessionStub := &sessionManagerStub{
		DeleteFunc: func(ctx context.Context, sessionID string) error {
			deletedSessionID = sessionID
			return nil
		},
	}

	svc := service.NewAuthService(&mockRepo{}, sessionStub, time.Minute)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.Logout(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "abc123", deletedSessionID)
	require.Contains(t, resp.Header.Get("Set-Cookie"), "session_id=")
	require.Contains(t, resp.Header.Get("Set-Cookie"), "Max-Age=")

}

func TestRegisterUser_MethodNotAllowed(t *testing.T) {
	svc := service.NewAuthService(&mockRepo{}, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	h.RegisterUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestRegisterUser_InvalidJSON(t *testing.T) {
	svc := service.NewAuthService(&mockRepo{}, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{invalid json"))
	w := httptest.NewRecorder()

	h.RegisterUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegisterUser_InternalServerError(t *testing.T) {
	repo := &mockRepo{
		GetByLoginFunc: func(login string) (*user.User, error) {
			return nil, nil
		},
		CreateFunc: func(u *user.User) error {
			return errors.New("db write error")
		},
	}
	svc := service.NewAuthService(repo, &sessionManagerStub{}, time.Minute)
	h := NewAuthHandler(svc)

	body := []byte(`{"name":"John","login":"john","password":"1234"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
