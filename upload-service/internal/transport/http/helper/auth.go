package auth

import (
	"net/http"
	"strconv"

	appredis "upload-service/internal/infrastructure/redis"

	redislib "github.com/go-redis/redis/v8"
)

func GetCurrentArtistID(r *http.Request, redisClient *redislib.Client) (int64, error) {
	sessionID := r.Header.Get("Authorization")
	if sessionID == "" {
		return 0, ErrUnauthorized
	}

	userID, err := appredis.GetUserID(redisClient, sessionID)
	if err != nil {
		return 0, ErrUnauthorized
	}

	artistIDStr, err := appredis.GetArtistID(redisClient, userID)
	if err != nil {
		return 0, ErrForbidden
	}

	artistID, err := strconv.ParseInt(artistIDStr, 10, 64)
	if err != nil {
		return 0, ErrInternal
	}

	return artistID, nil
}

func GetUserID(r *http.Request, redisClient *redislib.Client) (int64, string, error) {
	sessionID := r.Header.Get("Authorization")
	if sessionID == "" {
		return 0, "", ErrUnauthorized
	}

	userIDStr, err := appredis.GetUserID(redisClient, sessionID)
	if err != nil {
		return 0, "", ErrUnauthorized
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, "", ErrInternal
	}

	return userID, userIDStr, nil
}

var (
	ErrUnauthorized = &HttpError{Code: http.StatusUnauthorized, Message: "Неавторизованный доступ"}
	ErrForbidden    = &HttpError{Code: http.StatusForbidden, Message: "Доступ запрещён"}
	ErrInternal     = &HttpError{Code: http.StatusInternalServerError, Message: "Ошибка сервера"}
)

type HttpError struct {
	Code    int
	Message string
}

func (e *HttpError) Error() string {
	return e.Message
}
