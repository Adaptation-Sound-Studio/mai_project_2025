package service

import (
	"auth_service/internal/domain/user"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo user.UserRepository
}

func NewAuthService(repo user.UserRepository) *AuthService {
	return &AuthService{Repo: repo}
}

func (s *AuthService) RegisterUser(name, login, password string) error {
	existingUser, err := s.Repo.GetByLogin(login)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser := &user.User{
		Name:      name,
		Login:     login,
		Pass:      string(hashedPass),
		IsDeleted: false,
	}

	err = s.Repo.Create(newUser)
	if err != nil {
		return err
	}

	return nil
}
