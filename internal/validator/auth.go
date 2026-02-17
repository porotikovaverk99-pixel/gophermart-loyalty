package validator

import (
	"errors"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
)

var (
	ErrInvalidInput    = errors.New("login and password are required")
	ErrInvalidLogin    = errors.New("login must be 3-50 characters")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")
)

// ValidateAuth проверяет корректность логина и пароля.
func ValidateAuth(reqs model.RequestAuth) error {
	if reqs.Login == "" || reqs.Password == "" {
		return ErrInvalidInput
	}

	if len(reqs.Login) < 3 || len(reqs.Login) > 50 {
		return ErrInvalidLogin
	}

	if len(reqs.Password) < 6 {
		return ErrInvalidPassword
	}

	return nil
}
