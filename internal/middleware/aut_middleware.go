package middleware

import (
	"context"
	"errors"
	"fmt"
	"med_book/internal/service"
	"net/http"

	"github.com/google/uuid"
)

type (
	contextKey string
	UserID     string
)

const (
	UserIDContextKey    contextKey = "user_id"
	AuthTokenCookieName string     = "auth_token"
)

func AuthMiddleware(authService *service.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := extractTokenFromCookie(r, AuthTokenCookieName)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				// http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			userID, err := authService.GetUserIDFromToken(tokenString)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				// http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractTokenFromCookie извлекает значение токена из cookie
func extractTokenFromCookie(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("cookie %s: %w", cookieName, err)
	}

	if cookie.Value == "" {
		return "", errors.New("empty authentication token")
	}

	return cookie.Value, nil
}

// GetUserIDFromContext извлекает UUID пользователя из контекста
// Возвращает uuid.UUID и флаг успеха
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}
