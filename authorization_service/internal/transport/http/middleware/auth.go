package middleware

import (
	"auth_service/internal/service"
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

func AuthMiddleware(sessionService *service.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Session ID not found in cookies", http.StatusUnauthorized)
				return
			}
			sessionID := cookie.Value

			userID, err := sessionService.GetUserIDFromSession(r.Context(), sessionID)
			if err != nil {
				http.Error(w, "Сессия не найдена или истекла", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
