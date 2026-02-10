package handler

import (
	"encoding/json"
	"net/http"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
)

type BalanceHandler struct {
	service *service.BalanceService
}

func NewBalanceHandler(service *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		service: service,
	}
}

func (h *BalanceHandler) GetBalanceHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value("userID").(int64)
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

func (h *BalanceHandler) BalanceWithdrawHandler() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value("userID").(int64)
		if !ok {
			jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		defer r.Body.Close()

		var reqs model.ResponseWithdraw

		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			jsonError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := h.service.Withdraw(r.Context(), reqs, userID)

		if err != nil {
			switch {

			default:
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)

	})
}
