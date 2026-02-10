package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

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
			case errors.Is(err, service.ErrInvalidInput):
				jsonError(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidLogin):
				jsonError(w, err.Error(), http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidPassword):
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
