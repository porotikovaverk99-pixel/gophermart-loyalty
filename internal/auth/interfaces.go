// Package auth предоставляет интерфейсы для аутентификации.
package auth

// contextKey — собственный тип для ключей контекста.
type contextKey string

const (
	// UserIDKey — ключ для хранения ID пользователя в контексте.
	UserIDKey contextKey = "userID"
	// UserLoginKey — ключ для хранения логина пользователя в контексте.
	UserLoginKey contextKey = "userLogin"
)

// Manager определяет контракт для работы с токенами аутентификации.
type Manager interface {
	// Generate создает токен для пользователя.
	Generate(userID int64, login string) (string, error)

	// Validate проверяет токен и возвращает информацию о пользователе.
	Validate(tokenString string) (*UserInfo, error)
}

// UserInfo содержит данные пользователя из токена.
type UserInfo struct {
	UserID int64
	Login  string
}
