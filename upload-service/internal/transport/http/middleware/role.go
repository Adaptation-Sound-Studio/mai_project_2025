package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	appredis "upload-service/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
)

type contextKey string

const roleKey contextKey = "role"

func RequireAnyRole(redisClient *redis.Client, allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, role := range allowedRoles {
		allowed[strings.ToLower(strings.TrimSpace(role))] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Session ID not found in cookies", http.StatusUnauthorized)
				return
			}
			sessionID := cookie.Value

			if err != nil {
				log.Printf("Middleware error: GetUserID failed for session_id %s: %v\n", sessionID, err)
				http.Error(w, "Сессия недействительна", http.StatusUnauthorized)
				return
			}

			if len(allowed) > 0 {
				our_role, err := appredis.GetUserRole(redisClient, sessionID)
				our_role = strings.ToLower(strings.TrimSpace(our_role))
				if err != nil {
					http.Error(w, "Не удалось определить роль пользователя", http.StatusForbidden)
					return
				}

				if !allowed[our_role] {
					http.Error(w, "Доступ запрещён", http.StatusForbidden)
					return
				}

				ctx := context.WithValue(r.Context(), roleKey, our_role)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx := context.WithValue(r.Context(), roleKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
