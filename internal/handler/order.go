package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

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

			userID, ok := r.Context().Value("userID").(int64)
			if !ok {
				jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			_, err = h.service.UploadOrder(r.Context(), userID, orderNumber)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrInvalidCredentials):
					jsonError(w, "invalid login or password", http.StatusUnauthorized)
				default:
					jsonError(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			w.WriteHeader(http.StatusOK)

		} else if r.Method == http.MethodGet {

			userID, ok := r.Context().Value("userID").(int64)
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
			} else {
				w.WriteHeader(http.StatusOK)
			}

			if err := json.NewEncoder(w).Encode(&result); err != nil {
				jsonError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

		} else {

			jsonError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return

		}

	})
}
