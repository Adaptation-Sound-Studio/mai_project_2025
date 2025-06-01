package service

import (
	"auth_service/internal/domain/user"
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo           user.UserRepository
	SessionService *SessionService
}

func NewUserService(repo user.UserRepository, sessionService *SessionService) *UserService {
	return &UserService{
		Repo:           repo,
		SessionService: sessionService,
	}
}

func (s *UserService) CreateUser(user *user.User) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Pass = string(hashedPass)
	return s.Repo.Create(user)
}

func (s *UserService) SoftDeleteUser(userID int64) error {
	user, err := s.Repo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.IsDeleted = true
	err = s.Repo.Update(user)
	if err != nil {
		return err
	}

	if err := s.SessionService.DeleteUserSessions(context.Background(), userID); err != nil {
		log.Printf("warning: failed to delete sessions for user %d: %v", userID, err)
	}

	return nil
}

func (s *UserService) GetUserByID(id int64) (*user.User, error) {
	return s.Repo.GetByID(id)
}

func (s *UserService) UpdateUser(user *user.User) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Pass = string(hashedPass)
	return s.Repo.Update(user)
}

func (s *UserService) UpdateUserRole(userID int64, role string) error {
	user, err := s.Repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить пользователя: %w", err)
	}
	if user == nil {
		return fmt.Errorf("пользователь с ID %d не найден", userID)
	}
	return s.Repo.UpdateUserRole(userID, role)
}
