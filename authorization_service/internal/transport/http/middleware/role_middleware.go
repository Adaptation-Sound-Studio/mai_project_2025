package middleware

import (
	"auth_service/internal/service"
	"context"
	"net/http"
)

const roleKey contextKey = "userRole"

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}

func hasRole(userRole string, allowedRoles []string) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

func RoleMiddleware(sessionService *service.SessionService, allowedRoles ...string) func(http.Handler) http.Handler {
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
				http.Error(w, "Сессия недействительна", http.StatusUnauthorized)
				return
			}

			roleRaw, err := sessionService.GetSessionField(r.Context(), sessionID, "role")
			if err != nil {
				http.Error(w, "Роль не найдена", http.StatusForbidden)
				return
			}

			roleStr, ok := roleRaw.(string)
			if !ok || !hasRole(roleStr, allowedRoles) {
				http.Error(w, "Доступ запрещён: недостаточно прав", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, roleKey, roleStr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
