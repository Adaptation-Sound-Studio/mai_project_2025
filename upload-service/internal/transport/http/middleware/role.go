package middleware

import (
	"context"
	"log"
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
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Session ID not found in cookies", http.StatusUnauthorized)
				return
			}
			sessionID := cookie.Value

			//userID, err := appredis.GetUserID(redisClient, r.Context(), sessionID)
			if err != nil {
				log.Printf("Middleware error: GetUserID failed for session_id %s: %v\n", sessionID, err)
				http.Error(w, "Сессия недействительна", http.StatusUnauthorized)
				return
			}

			// Если роли переданы, проверяем роль
			if len(allowed) > 0 {
				role, err := appredis.GetUserRole(redisClient, sessionID)
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
				return
			}

			// Если роли не переданы — просто передаем контекст дальше без роли
			ctx := context.WithValue(r.Context(), roleKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
