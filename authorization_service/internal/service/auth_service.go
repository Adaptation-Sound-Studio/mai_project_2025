package service

import (
	"auth_service/internal/domain/user"
	"auth_service/internal/infrastructure/session"
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo           user.UserRepository
	SessionManager session.SessionManager
	SessionTTL     time.Duration
}

func NewAuthService(repo user.UserRepository, sm session.SessionManager, ttl time.Duration) *AuthService {
	return &AuthService{
		Repo:           repo,
		SessionManager: sm,
		SessionTTL:     ttl,
	}
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

// LoginUser проверяет логин и пароль, создает сессию и возвращает ID сессии
func (s *AuthService) LoginUser(ctx context.Context, login, password string) (string, error) {
	usr, err := s.Repo.GetByLogin(login)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Login failed: no user with login %s", login)
			return "", errors.New("invalid login or password")
		}
		log.Printf("Login failed: error querying user by login %s: %v", login, err)
		return "", err
	}

	if usr.IsDeleted {
		log.Printf("Login failed: user %s is marked as deleted", login)
		return "", errors.New("user is deleted")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(usr.Pass), []byte(password))
	if err != nil {
		log.Printf("Login failed: incorrect password for login %s", login)
		return "", errors.New("invalid login or password")
	}

	// Генерация ID сессии
	sessionID := uuid.NewString()

	// Данные для сессии
	sessionData := map[string]interface{}{
		"user_id":    usr.ID,
		"is_deleted": usr.IsDeleted,
		"artist_id":  "", // пока пусто, будет обновлено другим сервисом
	}

	// Создание сессии в Redis с TTL
	err = s.SessionManager.CreateSession(ctx, sessionID, sessionData, s.SessionTTL)
	if err != nil {
		log.Printf("Login failed: error creating session for user %s (ID: %d): %v", login, usr.ID, err)
		return "", err
	}

	log.Printf("Login successful for user %s (ID: %d), session ID: %s", login, usr.ID, sessionID)

	return sessionID, nil
}

func (s *AuthService) LogoutUser(ctx context.Context, sessionID string) error {
	return s.SessionManager.DeleteSession(ctx, sessionID)
}
