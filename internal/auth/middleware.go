// Package auth предоставляет middleware для аутентификации через JWT-токены.
package auth

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет наличие и валидность токена в заголовке Authorization.
// Токен должен быть в формате: Bearer <token>
// При успешной валидации добавляет userID и userLogin в контекст запроса.
func AuthMiddleware(authManager Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			userInfo, err := authManager.Validate(token)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userInfo.UserID)
			ctx = context.WithValue(ctx, UserLoginKey, userInfo.Login)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
