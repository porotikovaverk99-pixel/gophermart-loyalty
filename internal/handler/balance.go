// Package handler обрабатывает HTTP-запросы и формирует ответы.
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
)

type BalanceService interface {
	GetUserBalance(ctx context.Context, userID int64) (model.BalanceResponse, error)
	CreateWithdraw(ctx context.Context, reqs model.WithdrawRequest, userID int64) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

// BalanceHandler обрабатывает запросы на получение баланса и списание баллов.
type BalanceHandler struct {
	service BalanceService
}

// NewBalanceHandler создает новый обработчик баланса.
func NewBalanceHandler(service BalanceService) *BalanceHandler {
	return &BalanceHandler{
		service: service,
	}
}

// GetBalanceHandler возвращает текущий баланс пользователя.
// GET /api/user/balance
// Headers: Authorization: Bearer <token>
// Success: 200 OK, {"current": 500.50, "withdrawn": 100.25}
// Errors: 401 Unauthorized, 500 Internal Server Error
func (h *BalanceHandler) GetBalanceHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(auth.UserIDKey).(int64)
		if !ok {
			jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		result, err := h.service.GetUserBalance(r.Context(), userID)

		if err != nil {
			switch {

			default:
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(&result); err != nil {
			jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	})
}

// BalanceWithdrawHandler списывает баллы с баланса пользователя.
// POST /api/user/balance/withdraw
// Headers: Authorization: Bearer <token>
// Body: {"order": "2377225624", "sum": 100.50}
// Success: 200 OK
// Errors:
//   - 400 Bad Request (неверный формат, отрицательная сумма)
//   - 401 Unauthorized
//   - 402 Payment Required (недостаточно средств)
//   - 409 Conflict (заказ уже использован)
//   - 422 Unprocessable Entity (невалидный номер заказа)
//   - 500 Internal Server Error
func (h *BalanceHandler) BalanceWithdrawHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(auth.UserIDKey).(int64)
		if !ok {
			jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		defer r.Body.Close()

		var reqs model.WithdrawRequest

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := h.service.CreateWithdraw(r.Context(), reqs, userID)

		if err != nil {
			switch err {
			case service.ErrInvalidAmount:
				jsonError(w, err.Error(), http.StatusBadRequest)
			case service.ErrInvalidOrderNumber:
				jsonError(w, err.Error(), http.StatusUnprocessableEntity)
			case service.ErrInsufficientFunds:
				jsonError(w, err.Error(), http.StatusPaymentRequired)
			case service.ErrOrderAlreadyWithdrawn:
				jsonError(w, err.Error(), http.StatusConflict)
			default:
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)

	})
}

// GetWithdrawalsHandler возвращает список списаний пользователя.
// GET /api/user/withdrawals
// Headers: Authorization: Bearer <token>
// Success: 200 OK + массив списаний, 204 No Content (нет списаний)
// Errors: 401 Unauthorized, 500 Internal Server Error
func (h *BalanceHandler) GetWithdrawalsHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(auth.UserIDKey).(int64)
		if !ok {
			jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		result, err := h.service.GetUserWithdrawals(r.Context(), userID)

		if err != nil {
			switch err {
			default:
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	})
}
