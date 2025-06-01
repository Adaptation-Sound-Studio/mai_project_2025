package service

import (
	"auth_service/internal/domain/user"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockSessionManager struct {
	mock.Mock
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockSessionManager) CreateSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	return m.Called(ctx, sessionID, data, expiration).Error(0)
}

func (m *MockSessionManager) UpdateSessionField(ctx context.Context, sessionID, field string, value interface{}) error {
	return m.Called(ctx, sessionID, field, value).Error(0)
}

func (m *MockSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.Called(ctx, sessionID).Error(0)
}

func (m *MockUserRepo) GetByLogin(login string) (*user.User, error) {
	args := m.Called(login)
	u := args.Get(0)
	if u == nil {
		return nil, args.Error(1)
	}
	return u.(*user.User), args.Error(1)
}

func (m *MockUserRepo) Create(u *user.User) error {
	return m.Called(u).Error(0)
}

func (m *MockUserRepo) GetByID(id int64) (*user.User, error) {
	args := m.Called(id)
	u := args.Get(0)
	if u == nil {
		return nil, args.Error(1)
	}
	return u.(*user.User), args.Error(1)
}

func (m *MockUserRepo) Update(u *user.User) error {
	return m.Called(u).Error(0)
}

func TestRegisterUser_Success(t *testing.T) {
	repo := new(MockUserRepo)
	sm := new(MockSessionManager)
	svc := NewAuthService(repo, sm, time.Hour)

	repo.On("GetByLogin", "testuser").Return(nil, nil)
	repo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	err := svc.RegisterUser("Test", "testuser", "password123")
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestRegisterUser_AlreadyExists(t *testing.T) {
	repo := new(MockUserRepo)
	sm := new(MockSessionManager)
	svc := NewAuthService(repo, sm, time.Hour)

	repo.On("GetByLogin", "existing").Return(&user.User{ID: 1}, nil)

	err := svc.RegisterUser("Any", "existing", "123")
	require.EqualError(t, err, "user already exists")
}

func TestLoginUser_Success(t *testing.T) {
	repo := new(MockUserRepo)
	sm := new(MockSessionManager)
	svc := NewAuthService(repo, sm, time.Hour)

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	existingUser := &user.User{
		ID:        42,
		Login:     "testuser",
		Pass:      string(hashedPass),
		IsDeleted: false,
	}

	repo.On("GetByLogin", "testuser").Return(existingUser, nil)
	sm.On("CreateSession", mock.Anything, mock.AnythingOfType("string"), mock.Anything, time.Hour).Return(nil)

	sessionID, err := svc.LoginUser(context.Background(), "testuser", "secret")
	require.NoError(t, err)
	require.NotEmpty(t, sessionID)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	repo := new(MockUserRepo)
	sm := new(MockSessionManager)
	svc := NewAuthService(repo, sm, time.Hour)

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &user.User{Pass: string(hashedPass)}

	repo.On("GetByLogin", "testuser").Return(user, nil)

	_, err := svc.LoginUser(context.Background(), "testuser", "wrong")
	require.EqualError(t, err, "invalid login or password")
}

func TestLoginUser_Deleted(t *testing.T) {
	repo := new(MockUserRepo)
	sm := new(MockSessionManager)
	svc := NewAuthService(repo, sm, time.Hour)

	repo.On("GetByLogin", "testuser").Return(&user.User{IsDeleted: true}, nil)

	_, err := svc.LoginUser(context.Background(), "testuser", "pass")
	require.EqualError(t, err, "user is deleted")
}

func TestLogoutUser(t *testing.T) {
	repo := new(MockUserRepo)
	sm := new(MockSessionManager)
	svc := NewAuthService(repo, sm, time.Hour)

	sm.On("DeleteSession", mock.Anything, "session-id").Return(nil)

	err := svc.LogoutUser(context.Background(), "session-id")
	require.NoError(t, err)
}
