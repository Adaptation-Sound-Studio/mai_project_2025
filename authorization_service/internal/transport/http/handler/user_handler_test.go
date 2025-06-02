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

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
	repo := &mockUserRepo{
		CreateFunc: func(u *user.User) error {
			u.ID = 1
			return nil
		},
	}
	svc := service.NewUserService(repo, nil)
	h := NewUserHandler(svc, nil)

	body := []byte(`{"name":"Alice","login":"alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.CreateUser(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestGetUserByID_Success(t *testing.T) {
	repo := &mockUserRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			return &user.User{ID: id, Name: "Test"}, nil
		},
	}
	svc := service.NewUserService(repo, nil)
	h := NewUserHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
	w := httptest.NewRecorder()

	h.GetUserByID(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateUserByID_Success(t *testing.T) {
	repo := &mockUserRepo{
		UpdateFunc: func(u *user.User) error {
			require.Equal(t, int64(5), u.ID)
			require.Equal(t, "Updated", u.Name)
			return nil
		},
	}
	svc := service.NewUserService(repo, nil)
	h := NewUserHandler(svc, nil)

	body := []byte(`{"name":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/5", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"user_id": "5"})
	w := httptest.NewRecorder()

	h.UpdateUserByID(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetCurrentUser_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			return &user.User{ID: id, Name: "Current"}, nil
		},
	}
	sessionRepo := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			require.Equal(t, "abc123", sessionID)
			return 7, nil
		},
	}
	userSvc := service.NewUserService(userRepo, nil)
	sessionSvc := service.NewSessionService(sessionRepo)

	h := NewUserHandler(userSvc, sessionSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateCurrentUser_Success(t *testing.T) {
	userRepo := &mockUserRepo{
		UpdateFunc: func(u *user.User) error {
			require.Equal(t, int64(10), u.ID)
			require.Equal(t, "NewName", u.Name)
			return nil
		},
	}
	sessionRepo := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 10, nil
		},
	}
	userSvc := service.NewUserService(userRepo, nil)
	sessionSvc := service.NewSessionService(sessionRepo)

	h := NewUserHandler(userSvc, sessionSvc)

	body := []byte(`{"name":"NewName"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sessX"})
	w := httptest.NewRecorder()

	h.UpdateCurrentUser(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateUserRole_Failure(t *testing.T) {
	mockUserRepo := &mockUserRepo{}
	mockSession := &mockSessionRepo{}

	// имитируем ошибку при UpdateUserRole
	mockUserRepo.GetByIDFunc = func(id int64) (*user.User, error) {
		return &user.User{ID: id, Name: "Test", Login: "test", Pass: "hash"}, nil
	}
	mockUserRepo.UpdateUserRoleFunc = func(userID int64, role string) error {
		return errors.New("update failed")
	}

	realSessionService := service.NewSessionService(mockSession)
	service := service.NewUserService(mockUserRepo, realSessionService)
	handler := NewUserHandler(service, realSessionService)

	body := []byte(`{"user_id":1,"role":"admin"}`)
	req := httptest.NewRequest(http.MethodPatch, "/users/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateUserRole(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to update user role")
}

func TestUpdateUserRole_Success(t *testing.T) {
	mockUserRepo := &mockUserRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			require.Equal(t, int64(1), id)
			return &user.User{
				ID:    1,
				Name:  "Test",
				Login: "test",
				Pass:  "hash",
			}, nil
		},
		UpdateUserRoleFunc: func(userID int64, role string) error {
			require.Equal(t, int64(1), userID)
			require.Equal(t, "moderator", role)
			return nil
		},
	}
	mockSession := &mockSessionRepo{}
	sessionService := service.NewSessionService(mockSession)
	userService := service.NewUserService(mockUserRepo, sessionService)
	handler := &UserHandler{Service: userService}

	body := bytes.NewReader([]byte(`{"user_id": 1, "role": "moderator"}`))
	req := httptest.NewRequest(http.MethodPost, "/users/role", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUserRole(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "User role updated")
}

func TestSoftDeleteUserByID_Success(t *testing.T) {
	mockUserRepo := &mockUserRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			require.Equal(t, int64(123), id)
			return &user.User{
				ID:        123,
				Name:      "Test",
				Login:     "test",
				Pass:      "pass",
				IsDeleted: false,
			}, nil
		},
		UpdateFunc: func(u *user.User) error {
			require.Equal(t, int64(123), u.ID)
			require.True(t, u.IsDeleted)
			return nil
		},
	}
	mockSession := &mockSessionRepo{
		DeleteSessionsByUserIDFunc: func(ctx context.Context, userID int64) error {
			require.Equal(t, int64(123), userID)
			return nil
		},
	}

	sessionService := service.NewSessionService(mockSession)
	userService := service.NewUserService(mockUserRepo, sessionService)
	handler := &UserHandler{Service: userService}

	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	req = mux.SetURLVars(req, map[string]string{"user_id": "123"})
	w := httptest.NewRecorder()

	handler.SoftDeleteUserByID(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestUpdateCurrentUser_NoSessionCookie(t *testing.T) {
	handler := &UserHandler{}

	req := httptest.NewRequest(http.MethodPut, "/update-me", nil)
	w := httptest.NewRecorder()

	handler.UpdateCurrentUser(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Session ID not found")
}

func TestUpdateCurrentUser_SessionNotFound(t *testing.T) {
	mockSession := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 0, errors.New("not found")
		},
	}
	handler := &UserHandler{SessionService: service.NewSessionService(mockSession)}

	req := httptest.NewRequest(http.MethodPut, "/update-me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	w := httptest.NewRecorder()

	handler.UpdateCurrentUser(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Сессия не найдена")
}
func TestUpdateCurrentUser_InvalidJSON(t *testing.T) {
	mockSession := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 42, nil
		},
	}

	sessionService := service.NewSessionService(mockSession)

	handler := &UserHandler{
		SessionService: sessionService,
	}

	body := strings.NewReader(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPut, "/update-me", body)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	w := httptest.NewRecorder()

	handler.UpdateCurrentUser(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "Неверный формат")
}

func TestUpdateCurrentUser_UpdateError(t *testing.T) {
	mockSession := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 42, nil
		},
	}

	mockUser := &mockUserRepo{
		UpdateFunc: func(u *user.User) error {
			return errors.New("update failed")
		},
	}

	sessionService := service.NewSessionService(mockSession)
	userService := service.NewUserService(mockUser, sessionService)

	handler := &UserHandler{
		SessionService: sessionService,
		Service:        userService,
	}

	body := strings.NewReader(`{"name": "Updated", "login": "newlogin", "password": "pass"}`)
	req := httptest.NewRequest(http.MethodPut, "/update-me", body)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	w := httptest.NewRecorder()

	handler.UpdateCurrentUser(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "Ошибка при обновлении")
}

func TestGetCurrentUser_MissingSessionCookie(t *testing.T) {
	handler := &UserHandler{}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Session ID not found")
}

func TestGetCurrentUser_SessionNotFound(t *testing.T) {
	sessionMock := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 0, errors.New("not found")
		},
	}
	sessionService := service.NewSessionService(sessionMock)

	handler := &UserHandler{
		SessionService: sessionService,
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "Сессия не найдена")
}

func TestGetCurrentUser_GetUserError(t *testing.T) {
	sessionMock := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 42, nil
		},
	}
	userMock := &mockUserRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			return nil, errors.New("db error")
		},
	}

	sessionService := service.NewSessionService(sessionMock)
	userService := service.NewUserService(userMock, sessionService)

	handler := &UserHandler{
		SessionService: sessionService,
		Service:        userService,
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "Ошибка при получении")
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	sessionMock := &mockSessionRepo{
		GetUserIDFromSessionFunc: func(ctx context.Context, sessionID string) (int64, error) {
			return 42, nil
		},
	}
	userMock := &mockUserRepo{
		GetByIDFunc: func(id int64) (*user.User, error) {
			return nil, nil
		},
	}

	sessionService := service.NewSessionService(sessionMock)
	userService := service.NewUserService(userMock, sessionService)

	handler := &UserHandler{
		SessionService: sessionService,
		Service:        userService,
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "Пользователь не найден")
}

func TestUpdateUserByID_InvalidUserID(t *testing.T) {
	handler := &UserHandler{}

	req := httptest.NewRequest(http.MethodPut, "/users/abc", strings.NewReader(`{}`))
	req = mux.SetURLVars(req, map[string]string{"user_id": "abc"})
	w := httptest.NewRecorder()

	handler.UpdateUserByID(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "Неверный user_id")
}

func TestUpdateUserByID_InvalidJSON(t *testing.T) {
	handler := &UserHandler{}

	req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(`{invalid json}`))
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
	w := httptest.NewRecorder()

	handler.UpdateUserByID(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "Неверный формат запроса")
}

func TestUpdateUserByID_ServiceError(t *testing.T) {
	mockUser := &mockUserRepo{
		UpdateFunc: func(u *user.User) error {
			return errors.New("update failed")
		},
	}
	mockSession := &mockSessionRepo{}
	svc := service.NewUserService(mockUser, service.NewSessionService(mockSession))
	handler := &UserHandler{Service: svc}

	body := `{"name": "New", "login": "newlogin", "password": "newpass"}`
	req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"user_id": "1"})
	w := httptest.NewRecorder()

	handler.UpdateUserByID(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "Ошибка при обновлении пользователя")
}
