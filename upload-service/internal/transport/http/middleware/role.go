package middleware

import (
	"context"
	"net/http"

	appredis "upload-service/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
)

type contextKey string

const roleKey contextKey = "role"

func RequireAnyRole(redisClient *redis.Client, allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, role := range allowedRoles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := r.Header.Get("Authorization")
			if sessionID == "" {
				http.Error(w, "Не указан токен авторизации", http.StatusUnauthorized)
				return
			}

			userID, err := appredis.GetUserID(redisClient, sessionID)
			if err != nil {
				http.Error(w, "Сессия недействительна", http.StatusUnauthorized)
				return
			}

			role, err := appredis.GetUserRole(redisClient, userID)
			if err != nil {
				http.Error(w, "Не удалось определить роль пользователя", http.StatusForbidden)
				return
			}

			if !allowed[role] {
				http.Error(w, "Доступ запрещён", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), roleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
