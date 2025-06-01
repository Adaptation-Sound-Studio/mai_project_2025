package service

import (
	"auth_service/internal/domain/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser(t *testing.T) {
	repo := new(MockUserRepo)
	service := NewUserService(repo)

	u := &user.User{Name: "Test"}
	repo.On("Create", u).Return(nil)

	err := service.CreateUser(u)
	require.NoError(t, err)
	repo.AssertCalled(t, "Create", u)
}

func TestUserService_GetUserByID(t *testing.T) {
	repo := new(MockUserRepo)
	service := NewUserService(repo)

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

func TestUserService_UpdateUser(t *testing.T) {
	repo := new(MockUserRepo)
	service := NewUserService(repo)

	u := &user.User{
		ID:        1,
		Name:      "Bob Updated",
		Login:     "bob",
		Pass:      "newhash",
		IsDeleted: false,
	}
	repo.On("Update", u).Return(nil)

	err := service.UpdateUser(u)
	require.NoError(t, err)
	repo.AssertCalled(t, "Update", u)
}
