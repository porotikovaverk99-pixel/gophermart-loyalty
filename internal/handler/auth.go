// Package handler обрабатывает HTTP-запросы и формирует ответы.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/validator"
)

type AuthService interface {
	Register(ctx context.Context, reqs model.RequestAuth) (string, error)
	Login(ctx context.Context, reqs model.RequestAuth) (string, error)
}

// AuthHandler обрабатывает запросы на регистрацию и аутентификацию.
type AuthHandler struct {
	service AuthService
}

// NewAuthHandler создает новый обработчик аутентификации.
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// RegisterHandler регистрирует нового пользователя.
// POST /api/user/register
// Body: {"login": "string", "password": "string"}
// Success: 200 OK, Authorization: Bearer <token>
// Errors: 400, 409, 500
func (h *AuthHandler) RegisterHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		var reqs model.RequestAuth

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		token, err := h.service.Register(r.Context(), reqs)
		if err != nil {
			switch {
			case errors.Is(err, validator.ErrInvalidInput):
				jsonError(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, validator.ErrInvalidLogin):
				jsonError(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, validator.ErrInvalidPassword):
				jsonError(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, service.ErrLoginAlreadyExists):
				jsonError(w, err.Error(), http.StatusConflict)
			default:
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)

	})
}

// LoginHandler аутентифицирует пользователя.
// POST /api/user/login
// Body: {"login": "string", "password": "string"}
// Success: 200 OK, Authorization: Bearer <token>
// Errors: 400, 401, 500
func (h *AuthHandler) LoginHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		var reqs model.RequestAuth

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		token, err := h.service.Login(r.Context(), reqs)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidCredentials):
				jsonError(w, "invalid login or password", http.StatusUnauthorized)
			default:
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)

	})
}
