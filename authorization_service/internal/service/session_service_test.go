package service

import (
	"auth_service/internal/domain/user"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSessionRepo struct {
	mock.Mock
}

func (m *MockSessionRepo) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	return m.Called(ctx, sessionID, data, expiration).Error(0)
}

func (m *MockSessionRepo) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	return m.Called(ctx, sessionID, field, value).Error(0)
}

func (m *MockSessionRepo) GetSessionField(ctx context.Context, sessionID string, field string) (interface{}, error) {
	args := m.Called(ctx, sessionID, field)
	return args.Get(0), args.Error(1)
}

func (m *MockSessionRepo) GetUserIDFromSession(ctx context.Context, sessionID string) (int64, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	return m.Called(ctx, sessionID).Error(0)
}

func (m *MockSessionRepo) DeleteSessionsByUserID(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

func TestSessionService_CreateSession(t *testing.T) {
	repo := new(MockSessionRepo)
	s := NewSessionService(repo)

	ctx := context.Background()
	sessionID := "sess-1"
	data := map[string]interface{}{"user_id": 1}
	exp := time.Minute

	repo.On("CreateSession", ctx, sessionID, data, exp).Return(nil)
	err := s.CreateSession(ctx, sessionID, data, exp)
	require.NoError(t, err)
	repo.AssertCalled(t, "CreateSession", ctx, sessionID, data, exp)
}

func TestSessionService_UpdateSessionField(t *testing.T) {
	repo := new(MockSessionRepo)
	s := NewSessionService(repo)

	ctx := context.Background()
	repo.On("UpdateSessionField", ctx, "sess-2", "role", "admin").Return(nil)

	err := s.UpdateSessionField(ctx, "sess-2", "role", "admin")
	require.NoError(t, err)
	repo.AssertCalled(t, "UpdateSessionField", ctx, "sess-2", "role", "admin")
}

func TestSessionService_GetSessionField(t *testing.T) {
	repo := new(MockSessionRepo)
	s := NewSessionService(repo)

	ctx := context.Background()
	repo.On("GetSessionField", ctx, "sess-3", "user_id").Return(int64(42), nil)

	val, err := s.GetSessionField(ctx, "sess-3", "user_id")
	require.NoError(t, err)
	require.Equal(t, int64(42), val)
	repo.AssertCalled(t, "GetSessionField", ctx, "sess-3", "user_id")
}

func TestSessionService_GetUserIDFromSession(t *testing.T) {
	repo := new(MockSessionRepo)
	s := NewSessionService(repo)

	ctx := context.Background()
	repo.On("GetUserIDFromSession", ctx, "sess-4").Return(int64(101), nil)

	id, err := s.GetUserIDFromSession(ctx, "sess-4")
	require.NoError(t, err)
	require.Equal(t, int64(101), id)
	repo.AssertCalled(t, "GetUserIDFromSession", ctx, "sess-4")
}

func TestSessionService_DeleteSession(t *testing.T) {
	repo := new(MockSessionRepo)
	s := NewSessionService(repo)

	ctx := context.Background()
	repo.On("DeleteSession", ctx, "sess-5").Return(nil)

	err := s.DeleteSession(ctx, "sess-5")
	require.NoError(t, err)
	repo.AssertCalled(t, "DeleteSession", ctx, "sess-5")
}

func TestSessionService_DeleteUserSessions(t *testing.T) {
	repo := new(MockSessionRepo)
	s := NewSessionService(repo)

	ctx := context.Background()
	repo.On("DeleteSessionsByUserID", ctx, int64(555)).Return(nil)

	err := s.DeleteUserSessions(ctx, 555)
	require.NoError(t, err)
	repo.AssertCalled(t, "DeleteSessionsByUserID", ctx, int64(555))
}

func TestUserService_SoftDeleteUser_NotFound(t *testing.T) {
	repo := new(MockUserRepo)
	mockRepo := new(MockSessionRepo)
	session := NewSessionService(mockRepo)
	service := NewUserService(repo, session)

	repo.On("GetByID", int64(404)).Return(nil, nil)

	err := service.SoftDeleteUser(404)
	require.Error(t, err)
	require.Equal(t, "user not found", err.Error())
}

func TestUserService_UpdateUserRole(t *testing.T) {
	repo := new(MockUserRepo)
	mockSessionRepo := new(MockSessionRepo)
	session := NewSessionService(mockSessionRepo)
	service := NewUserService(repo, session)

	u := &user.User{ID: 77}
	repo.On("GetByID", int64(77)).Return(u, nil)
	repo.On("UpdateUserRole", int64(77), "editor").Return(nil)

	err := service.UpdateUserRole(77, "editor")
	require.NoError(t, err)
	repo.AssertCalled(t, "UpdateUserRole", int64(77), "editor")
}

func TestUserService_UpdateUserRole_NotFound(t *testing.T) {
	repo := new(MockUserRepo)
	mockSessionRepo := new(MockSessionRepo)
	session := NewSessionService(mockSessionRepo)
	service := NewUserService(repo, session)

	repo.On("GetByID", int64(404)).Return(nil, nil)

	err := service.UpdateUserRole(404, "admin")
	require.Error(t, err)
	require.EqualError(t, err, "пользователь с ID 404 не найден")
}

func TestUserService_SoftDeleteUser(t *testing.T) {
	repo := new(MockUserRepo)
	mockRepo := new(MockSessionRepo)
	session := NewSessionService(mockRepo)
	service := NewUserService(repo, session)

	u := &user.User{ID: 100}
	repo.On("GetByID", int64(100)).Return(u, nil)
	repo.On("Update", u).Return(nil)
	mockRepo.On("DeleteSessionsByUserID", mock.Anything, int64(100)).Return(nil)

	err := service.SoftDeleteUser(100)
	require.NoError(t, err)
	repo.AssertCalled(t, "Update", u)
	mockRepo.AssertCalled(t, "DeleteSessionsByUserID", mock.Anything, int64(100))
}
