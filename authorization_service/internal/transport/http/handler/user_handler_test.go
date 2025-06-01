package handler

import (
	"auth_service/internal/domain/user"
	"auth_service/internal/infrastructure/session"
	"auth_service/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	CreateFunc     func(*user.User) error
	GetByIDFunc    func(int64) (*user.User, error)
	UpdateFunc     func(*user.User) error
	GetByLoginFunc func(string) (*user.User, error)
}

func (m *mockRepo) Create(u *user.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(u)
	}
	return nil
}
func (m *mockRepo) GetByID(id int64) (*user.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}
func (m *mockRepo) Update(u *user.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(u)
	}
	return nil
}
func (m *mockRepo) GetByLogin(login string) (*user.User, error) {
	if m.GetByLoginFunc != nil {
		return m.GetByLoginFunc(login)
	}
	return nil, nil
}

type sessionManagerStub struct {
	DeleteFunc    func(ctx context.Context, sessionID string) error
	GetUserIDFunc func(ctx context.Context, sessionID string) (int64, error)
}

func (s *sessionManagerStub) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	return nil
}

func (s *sessionManagerStub) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	return nil
}

func (s *sessionManagerStub) DeleteSession(ctx context.Context, sessionID string) error {
	if s.DeleteFunc != nil {
		return s.DeleteFunc(ctx, sessionID)
	}
	return nil
}

func (s *sessionManagerStub) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
	if s.GetUserIDFunc != nil {
		return s.GetUserIDFunc(ctx, sessionID)
	}
	return 0, nil
}
func TestCreateUser_Success(t *testing.T) {
	var createdUser *user.User

	repo := &mockRepo{
		CreateFunc: func(u *user.User) error {
			createdUser = u
			return nil
		},
	}

	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, nil)

	body := []byte(`{"name":"Alice","login":"alice","password":"secure"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var got user.User
	err := json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	require.Equal(t, "Alice", got.Name)
	require.Equal(t, "alice", got.Login)
	require.Equal(t, "secure", got.Pass)
	require.Equal(t, *createdUser, got)
}

func TestGetUserByID_Success(t *testing.T) {
	expectedUser := &user.User{
		ID:    123,
		Name:  "Alice",
		Login: "alice",
		Pass:  "secure",
	}

	repo := &mockRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			require.Equal(t, int64(123), id)
			return expectedUser, nil
		},
	}

	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	req = mux.SetURLVars(req, map[string]string{"user_id": "123"})

	w := httptest.NewRecorder()
	h.GetUserByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got user.User
	err := json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)
	require.Equal(t, *expectedUser, got)
}

func TestGetCurrentUser_Success(t *testing.T) {
	expectedUser := &user.User{
		ID:    7,
		Name:  "Vlad",
		Login: "vlad",
		Pass:  "hash",
	}

	repo := &mockRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			require.Equal(t, int64(7), id)
			return expectedUser, nil
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	sessions := session.NewRedisSessionManager(client)

	err = client.Set(context.Background(), "abc123", `{"user_id":7}`, time.Minute).Err()
	require.NoError(t, err)

	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got user.User
	err = json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)
	require.Equal(t, *expectedUser, got)
}

func TestUpdateCurrentUser_Success(t *testing.T) {
	expectedUser := &user.User{
		ID:    7,
		Name:  "Updated Vlad",
		Login: "vlad_updated",
		Pass:  "newpass",
	}

	var updatedUser *user.User

	repo := &mockRepo{
		UpdateFunc: func(u *user.User) error {
			updatedUser = u
			return nil
		},
	}

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	err = client.Set(context.Background(), "abc123", `{"user_id":7}`, time.Minute).Err()
	require.NoError(t, err)

	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, sessions)

	body := []byte(`{"name":"Updated Vlad","login":"vlad_updated","password":"newpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/users/me", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.UpdateCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got user.User
	err = json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	require.Equal(t, *expectedUser, got)
	require.Equal(t, *expectedUser, *updatedUser)
}

func TestUpdateCurrentUser_MissingSessionCookie(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
	w := httptest.NewRecorder()

	h.UpdateCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Session ID not found")
}

func TestUpdateCurrentUser_SessionNotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.UpdateCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Сессия не найдена")
}

func TestUpdateCurrentUser_InvalidJSON(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	err = client.Set(context.Background(), "abc123", `{"user_id":7}`, time.Minute).Err()
	require.NoError(t, err)

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodPost, "/users/me", bytes.NewReader([]byte("{invalid_json")))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.UpdateCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Неверный формат")
}

func TestUpdateCurrentUser_UpdateError(t *testing.T) {
	repo := &mockRepo{
		UpdateFunc: func(u *user.User) error {
			return errors.New("database error")
		},
	}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	err = client.Set(context.Background(), "abc123", `{"user_id":7}`, time.Minute).Err()
	require.NoError(t, err)

	h := NewUserHandler(svc, sessions)

	body := []byte(`{"name":"Fail","login":"fail","password":"err"}`)
	req := httptest.NewRequest(http.MethodPost, "/users/me", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.UpdateCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Ошибка при обновлении")
}

func TestGetCurrentUser_MissingSessionCookie(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Session ID not found")
}

func TestGetCurrentUser_SessionNotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	// Сессия "abc123" не записана в Redis

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Сессия не найдена")
}

func TestGetCurrentUser_RepoError(t *testing.T) {
	repo := &mockRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			return nil, errors.New("db down")
		},
	}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	err = client.Set(context.Background(), "abc123", `{"user_id":42}`, time.Minute).Err()
	require.NoError(t, err)

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Ошибка при получении пользователя")
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	repo := &mockRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			return nil, nil // пользователь не найден
		},
	}
	svc := service.NewUserService(repo)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sessions := session.NewRedisSessionManager(client)

	err = client.Set(context.Background(), "abc123", `{"user_id":42}`, time.Minute).Err()
	require.NoError(t, err)

	h := NewUserHandler(svc, sessions)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Пользователь не найден")
}

func TestUpdateUserByID_InvalidUserID(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, nil)

	req := httptest.NewRequest(http.MethodPut, "/users/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"user_id": "abc"})

	w := httptest.NewRecorder()
	h.UpdateUserByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Неверный user_id")
}

func TestUpdateUserByID_InvalidJSON(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, nil)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewReader([]byte("{invalid")))
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})

	w := httptest.NewRecorder()
	h.UpdateUserByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Неверный формат запроса")
}

func TestUpdateUserByID_UpdateError(t *testing.T) {
	repo := &mockRepo{
		UpdateFunc: func(u *user.User) error {
			return errors.New("fail to update")
		},
	}
	svc := service.NewUserService(repo)
	h := NewUserHandler(svc, nil)

	body := []byte(`{"name":"test","login":"test","password":"123"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/5", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"user_id": "5"})

	w := httptest.NewRecorder()
	h.UpdateUserByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, w.Body.String(), "Ошибка при обновлении пользователя")
}
