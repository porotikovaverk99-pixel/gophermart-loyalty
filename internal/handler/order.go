// Package handler обрабатывает HTTP-запросы и формирует ответы.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
)

type OrderService interface {
	UploadOrder(ctx context.Context, userID int64, number string) (int64, error)
	GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error)
}

// OrderHandler обрабатывает запросы на загрузку и получение заказов.
type OrderHandler struct {
	service OrderService
}

// NewOrderHandler создает новый обработчик заказов.
func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

// BaseOrderHandler обрабатывает POST и GET запросы для заказов.
//
// POST /api/user/orders
// Headers: Authorization: Bearer <token>
// Body: номер заказа (простая строка, не JSON)
// Success: 200 OK (заказ уже загружен), 202 Accepted (новый заказ принят)
// Errors: 400 Bad Request, 401 Unauthorized, 409 Conflict (заказ загружен другим пользователем),
//
//	422 Unprocessable Entity (невалидный номер), 500 Internal Server Error
//
// GET /api/user/orders
// Headers: Authorization: Bearer <token>
// Success: 200 OK + массив заказов, 204 No Content (нет заказов)
// Errors: 401 Unauthorized, 500 Internal Server Error
func (h *OrderHandler) BaseOrderHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {

			defer r.Body.Close()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			orderNumber := strings.TrimSpace(string(body))
			if orderNumber == "" {
				jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			userID, ok := r.Context().Value(auth.UserIDKey).(int64)
			if !ok {
				jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			_, err = h.service.UploadOrder(r.Context(), userID, orderNumber)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrInvalidOrderNumber):
					jsonError(w, err.Error(), http.StatusUnprocessableEntity)
				case errors.Is(err, service.ErrOrderBelongsToAnother):
					jsonError(w, err.Error(), http.StatusConflict)
				case errors.Is(err, service.ErrNumberAlreadyExists):
					w.WriteHeader(http.StatusOK)
					return
				default:
					jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}

			w.WriteHeader(http.StatusAccepted)

		} else if r.Method == http.MethodGet {

			userID, ok := r.Context().Value(auth.UserIDKey).(int64)
			if !ok {
				jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			result, err := h.service.GetUserOrders(r.Context(), userID)
			if err != nil {
				switch {
				default:
					jsonError(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")

			if len(result) == 0 {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.WriteHeader(http.StatusOK)

			if err := json.NewEncoder(w).Encode(&result); err != nil {
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

		} else {

			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return

		}

	})
}
