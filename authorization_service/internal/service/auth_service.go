package service

import (
	"auth_service/internal/domain/session"
	"auth_service/internal/domain/user"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo           user.UserRepository
	SessionManager session.SessionRepository
	SessionTTL     time.Duration
	uploadBaseURL  string
	apiKey         string
}

func NewAuthService(repo user.UserRepository, sm session.SessionRepository, ttl time.Duration, uploadBaseURL, apiKey string) *AuthService {
	return &AuthService{
		Repo:           repo,
		SessionManager: sm,
		SessionTTL:     ttl,
		uploadBaseURL:  uploadBaseURL,
		apiKey:         apiKey,
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

	err = bcrypt.CompareHashAndPassword([]byte(usr.Pass), []byte(password))
	if err != nil {
		log.Printf("Login failed: incorrect password for login %s", login)
		return "", errors.New("invalid login or password")
	}

	role, err := s.Repo.GetRoleByUserID(usr.ID)
	if err != nil {
		log.Printf("Login failed: error fetching role for user %s (ID: %d): %v", login, usr.ID, err)
		return "", err
	}

	var artistID string
	{
		url := fmt.Sprintf("%s/artists/user/%d", s.uploadBaseURL, usr.ID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Login: ошибка создания HTTP-запроса в upload-сервис: %v", err)
		} else {
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", s.apiKey)

			resp, err := http.DefaultClient.Do(req) // используем стандартный клиент
			if err != nil {
				log.Printf("Login: ошибка при запросе в upload-сервис: %v", err)
			} else {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var res struct {
						ArtistID string `json:"artist_id"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
						log.Printf("Login: ошибка декодирования ответа upload-сервиса: %v", err)
					} else {
						artistID = res.ArtistID
					}
				} else {
					log.Printf("Login: upload-сервис вернул статус %d", resp.StatusCode)
				}
			}
		}
	}

	sessionID := uuid.NewString()

	sessionData := map[string]interface{}{
		"user_id":    usr.ID,
		"is_deleted": usr.IsDeleted,
		"artist_id":  artistID,
		"role":       role,
	}

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
