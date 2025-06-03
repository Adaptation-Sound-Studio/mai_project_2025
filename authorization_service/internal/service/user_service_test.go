package service

import (
	"auth_service/internal/domain/user"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	repo := new(MockUserRepo)
	dummySession := &SessionService{}
	service := NewUserService(repo, dummySession)

	u := &user.User{Name: "Test"}
	repo.On("Create", u).Return(nil)

	err := service.CreateUser(u)
	require.NoError(t, err)
	repo.AssertCalled(t, "Create", u)
}

func TestUserService_GetUserByID(t *testing.T) {
	repo := new(MockUserRepo)
	dummySession := &SessionService{}
	service := NewUserService(repo, dummySession)

	expectedUser := &user.User{
		ID:    42,
		Name:  "Alice",
		Login: "alice",
		Pass:  "hashedpass",
	}
	repo.On("GetByID", int64(42)).Return(expectedUser, nil)

	result, err := service.GetUserByID(42)
	require.NoError(t, err)
	require.Equal(t, expectedUser, result)
	repo.AssertCalled(t, "GetByID", int64(42))
}

func TestSoftDeleteUser_WarningOnSessionError(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockSession := new(MockSessionManager)

	sessionService := NewSessionService(mockSession)

	svc := NewUserService(mockRepo, sessionService)

	mockRepo.On("GetByID", int64(123)).Return(&user.User{
		ID:        123,
		Name:      "Alice",
		Login:     "alice",
		Pass:      "hash",
		IsDeleted: false,
	}, nil)

	mockRepo.On("Update", mock.MatchedBy(func(u *user.User) bool {
		return u.ID == 123 && u.IsDeleted
	})).Return(nil)

	mockSession.On("DeleteSessionsByUserID", mock.Anything, int64(123)).
		Return(errors.New("redis connection failed"))

	err := svc.SoftDeleteUser(123)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestUpdateUserRole_GetByIDError(t *testing.T) {
	mockRepo := new(MockUserRepo)
	sessionService := NewSessionService(new(MockSessionManager))

	svc := NewUserService(mockRepo, sessionService)

	mockRepo.On("GetByID", int64(100)).Return(nil, errors.New("database error"))

	err := svc.UpdateUserRole(100, "moderator")

	require.Error(t, err)
	require.Contains(t, err.Error(), "не удалось получить пользователя")
	require.Contains(t, err.Error(), "database error")

	mockRepo.AssertCalled(t, "GetByID", int64(100))
	mockRepo.AssertNotCalled(t, "UpdateUserRole", mock.Anything, mock.Anything)
}
