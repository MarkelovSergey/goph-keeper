package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/MarkelovSergey/goph-keeper/internal/server/service"
)

// contextKey — тип для ключей контекста запроса.
type contextKey string

// UserIDKey — ключ для хранения ID пользователя в контексте.
const UserIDKey contextKey = "user_id"

// Auth проверяет JWT-токен в заголовке Authorization и добавляет ID пользователя в контекст.
func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "требуется авторизация", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			userID, err := authSvc.ParseToken(tokenStr)
			if err != nil {
				http.Error(w, "невалидный токен", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID извлекает ID пользователя из контекста запроса.
func GetUserID(r *http.Request) (uuid.UUID, bool) {
	id, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	return id, ok
}
