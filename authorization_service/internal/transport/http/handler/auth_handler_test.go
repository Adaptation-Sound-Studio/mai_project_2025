package handler

import (
	"auth_service/internal/domain/user"
	"auth_service/internal/service"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	CreateFunc          func(*user.User) error
	GetByLoginFunc      func(string) (*user.User, error)
	GetRoleByUserIDFunc func(int64) (string, error)
	GetByIDFunc         func(int64) (*user.User, error)
	UpdateFunc          func(*user.User) error
	UpdateUserRoleFunc  func(int64, string) error
	SoftDeleteUserFunc  func(int64) error
}

func (m *mockUserRepo) Create(u *user.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(u)
	}
	return nil
}

func (m *mockUserRepo) GetByLogin(login string) (*user.User, error) {
	if m.GetByLoginFunc != nil {
		return m.GetByLoginFunc(login)
	}
	return nil, nil
}

func (m *mockUserRepo) GetRoleByUserID(userID int64) (string, error) {
	if m.GetRoleByUserIDFunc != nil {
		return m.GetRoleByUserIDFunc(userID)
	}
	return "", nil
}

func (m *mockUserRepo) GetByID(id int64) (*user.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(u *user.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(u)
	}
	return nil
}

func (m *mockUserRepo) UpdateUserRole(userID int64, role string) error {
	if m.UpdateUserRoleFunc != nil {
		return m.UpdateUserRoleFunc(userID, role)
	}
	return nil
}

type mockSessionRepo struct {
	CreateSessionFunc          func(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error
	GetUserIDFromSessionFunc   func(ctx context.Context, sessionID string) (int64, error)
	GetSessionFieldFunc        func(ctx context.Context, sessionID string, field string) (interface{}, error)
	UpdateSessionFieldFunc     func(ctx context.Context, sessionID, field string, value interface{}) error
	DeleteSessionFunc          func(ctx context.Context, sessionID string) error
	DeleteSessionsByUserIDFunc func(ctx context.Context, userID int64) error
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, sessionID, data, expiration)
	}
	return nil
}

func (m *mockSessionRepo) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
	if m.GetUserIDFromSessionFunc != nil {
		return m.GetUserIDFromSessionFunc(ctx, sessionID)
	}
	return 0, errors.New("not implemented")
}

func (m *mockSessionRepo) GetSessionField(ctx context.Context, sessionID string, field string) (interface{}, error) {
	if m.GetSessionFieldFunc != nil {
		return m.GetSessionFieldFunc(ctx, sessionID, field)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepo) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	if m.UpdateSessionFieldFunc != nil {
		return m.UpdateSessionFieldFunc(ctx, sessionID, field, value)
	}
	return nil
}

func (m *mockSessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *mockSessionRepo) DeleteSessionsByUserID(ctx context.Context, userID int64) error {
	if m.DeleteSessionsByUserIDFunc != nil {
		return m.DeleteSessionsByUserIDFunc(ctx, userID)
	}
	return nil
}

const (
	hashSalt   = "test-salt"
	signingKey = "test-signing-key"
)

func TestRegisterUser_Success(t *testing.T) {
	repo := &mockUserRepo{
		GetByLoginFunc: func(login string) (*user.User, error) { return nil, nil },
		CreateFunc:     func(u *user.User) error { return nil },
	}
	sessionRepo := &mockSessionRepo{}
	svc := service.NewAuthService(repo, sessionRepo, time.Minute, hashSalt, signingKey)
	h := NewAuthHandler(svc)

	body := []byte(`{"name":"John","login":"john","password":"1234"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.RegisterUser(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestRegisterUser_InvalidJSON(t *testing.T) {
	svc := service.NewAuthService(&mockUserRepo{}, &mockSessionRepo{}, time.Minute, hashSalt, signingKey)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{bad json`)))
	w := httptest.NewRecorder()

	h.RegisterUser(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	password := "1234"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	repo := &mockUserRepo{
		GetByLoginFunc: func(login string) (*user.User, error) {
			return &user.User{ID: 1, Login: login, Pass: string(hash)}, nil
		},
	}
	sessionRepo := &mockSessionRepo{
		CreateSessionFunc: func(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
			return nil
		},
	}
	svc := service.NewAuthService(repo, sessionRepo, time.Minute, hashSalt, signingKey)
	h := NewAuthHandler(svc)

	body := []byte(`{"login":"john","password":"1234"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Set-Cookie"), "session_id=")
}

func TestLogin_InvalidJSON(t *testing.T) {
	svc := service.NewAuthService(&mockUserRepo{}, &mockSessionRepo{}, time.Minute, hashSalt, signingKey)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`bad json`)))
	w := httptest.NewRecorder()

	h.Login(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogout_MissingCookie(t *testing.T) {
	svc := service.NewAuthService(&mockUserRepo{}, &mockSessionRepo{}, time.Minute, hashSalt, signingKey)
	h := NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_Success(t *testing.T) {
	mockAuth := &service.AuthService{
		SessionManager: &mockSessionRepo{
			DeleteSessionFunc: func(ctx context.Context, sessionID string) error {
				require.Equal(t, "session123", sessionID)
				return nil
			},
		},
	}

	handler := &AuthHandler{AuthService: mockAuth}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "logout successful")
}

func TestLogout_DeleteSessionFails(t *testing.T) {
	mockAuth := &service.AuthService{
		SessionManager: &mockSessionRepo{
			DeleteSessionFunc: func(ctx context.Context, sessionID string) error {
				return errors.New("redis is down")
			},
		},
	}

	handler := &AuthHandler{AuthService: mockAuth}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "failed to logout")
}

func TestRegisterUser_WrongMethod(t *testing.T) {
	handler := &AuthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	rec := httptest.NewRecorder()

	handler.RegisterUser(rec, req)

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	require.Contains(t, rec.Body.String(), "Method Not Allowed")
}
