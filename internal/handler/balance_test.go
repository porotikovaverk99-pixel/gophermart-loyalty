package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/handler/mock"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestBalanceHandler_GetBalanceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		userID         interface{}
		setupMock      func(*mock.MockBalanceService)
		expectedStatus int
		expectedBody   *model.BalanceResponse
		expectedError  string
	}{
		{
			name:   "успешное получение баланса",
			method: http.MethodGet,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.GetBalanceResult = model.BalanceResponse{
					Current:   1500.50,
					Withdrawn: 300.25,
				}
				m.GetBalanceError = nil
			},
			expectedStatus: http.StatusOK,
			expectedBody: &model.BalanceResponse{
				Current:   1500.50,
				Withdrawn: 300.25,
			},
		},
		{
			name:   "пользователь не авторизован (нет userID в контексте)",
			method: http.MethodGet,
			userID: nil,
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name:   "пользователь не авторизован (userID не того типа)",
			method: http.MethodGet,
			userID: "not an int64",
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name:   "ошибка при получении баланса из сервиса",
			method: http.MethodGet,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.GetBalanceResult = model.BalanceResponse{}
				m.GetBalanceError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  http.StatusText(http.StatusInternalServerError),
		},
		{
			name:   "неверный метод (POST вместо GET)",
			method: http.MethodPost,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  http.StatusText(http.StatusMethodNotAllowed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockService := &mock.MockBalanceService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewBalanceHandler(mockService)

			req := httptest.NewRequest(tt.method, "/api/user/balance", nil)

			if tt.userID != nil {
				if uid, ok := tt.userID.(int64); ok {
					ctx := context.WithValue(req.Context(), auth.UserIDKey, uid)
					req = req.WithContext(ctx)
				}
			}

			w := httptest.NewRecorder()
			handler.GetBalanceHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp model.BalanceResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.Current, resp.Current)
				assert.Equal(t, tt.expectedBody.Withdrawn, resp.Withdrawn)
			} else if tt.expectedError != "" {
				var errResp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
				assert.Equal(t, tt.expectedError, errResp["error"])
			}
		})
	}
}

func TestBalanceHandler_BalanceWithdrawHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		userID         interface{}
		setupMock      func(*mock.MockBalanceService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "успешное списание",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "4111111111111111",
				Sum:   500,
			},
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.CreateWithdrawError = nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "неверный формат JSON",
			method:      http.MethodPost,
			requestBody: `{"order": "4111111111111111",`,
			userID:      int64(1),
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  http.StatusText(http.StatusBadRequest),
		},
		{
			name:        "пустое тело запроса",
			method:      http.MethodPost,
			requestBody: nil,
			userID:      int64(1),
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  http.StatusText(http.StatusBadRequest),
		},
		{
			name:   "отрицательная сумма",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "4111111111111111",
				Sum:   -100,
			},
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.CreateWithdrawError = service.ErrInvalidAmount
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  service.ErrInvalidAmount.Error(),
		},
		{
			name:   "невалидный номер заказа",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "123",
				Sum:   500,
			},
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.CreateWithdrawError = service.ErrInvalidOrderNumber
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  service.ErrInvalidOrderNumber.Error(),
		},
		{
			name:   "недостаточно средств",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "4111111111111111",
				Sum:   5000,
			},
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.CreateWithdrawError = service.ErrInsufficientFunds
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedError:  service.ErrInsufficientFunds.Error(),
		},
		{
			name:   "заказ уже использован",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "4111111111111111",
				Sum:   500,
			},
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.CreateWithdrawError = service.ErrOrderAlreadyWithdrawn
			},
			expectedStatus: http.StatusConflict,
			expectedError:  service.ErrOrderAlreadyWithdrawn.Error(),
		},
		{
			name:   "внутренняя ошибка сервиса",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "4111111111111111",
				Sum:   500,
			},
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.CreateWithdrawError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  http.StatusText(http.StatusInternalServerError),
		},
		{
			name:   "пользователь не авторизован",
			method: http.MethodPost,
			requestBody: model.WithdrawRequest{
				Order: "4111111111111111",
				Sum:   500,
			},
			userID: nil,
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name:        "неверный метод (GET)",
			method:      http.MethodGet,
			requestBody: nil,
			userID:      int64(1),
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  http.StatusText(http.StatusMethodNotAllowed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mock.MockBalanceService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewBalanceHandler(mockService)

			var req *http.Request
			if tt.requestBody != nil {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest(tt.method, "/api/user/balance/withdraw", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/api/user/balance/withdraw", nil)
			}

			if tt.userID != nil {
				if uid, ok := tt.userID.(int64); ok {
					ctx := context.WithValue(req.Context(), "userID", uid)
					req = req.WithContext(ctx)
				}
			}

			w := httptest.NewRecorder()
			handler.BalanceWithdrawHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK {
				var errResp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, errResp["error"])
				}
			}
		})
	}
}

func TestBalanceHandler_GetWithdrawalsHandler(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		method         string
		userID         interface{}
		setupMock      func(*mock.MockBalanceService)
		expectedStatus int
		expectedBody   []model.Withdrawal
		expectedError  string
	}{
		{
			name:   "успешное получение списка списаний",
			method: http.MethodGet,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.GetUserWithdrawalsResult = []model.Withdrawal{
					{
						Order:       "4111111111111111",
						Amount:      500,
						ProcessedAt: now,
					},
					{
						Order:       "5555555555554444",
						Amount:      300,
						ProcessedAt: now,
					},
				}
				m.GetUserWithdrawalsError = nil
			},
			expectedStatus: http.StatusOK,
			expectedBody: []model.Withdrawal{
				{
					Order:       "4111111111111111",
					Amount:      500,
					ProcessedAt: now,
				},
				{
					Order:       "5555555555554444",
					Amount:      300,
					ProcessedAt: now,
				},
			},
		},
		{
			name:   "пустой список списаний",
			method: http.MethodGet,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.GetUserWithdrawalsResult = []model.Withdrawal{}
				m.GetUserWithdrawalsError = nil
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   nil,
		},
		{
			name:   "пользователь не авторизован",
			method: http.MethodGet,
			userID: nil,
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name:   "внутренняя ошибка сервиса",
			method: http.MethodGet,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
				m.GetUserWithdrawalsResult = nil
				m.GetUserWithdrawalsError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  http.StatusText(http.StatusInternalServerError),
		},
		{
			name:   "неверный метод (POST)",
			method: http.MethodPost,
			userID: int64(1),
			setupMock: func(m *mock.MockBalanceService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  http.StatusText(http.StatusMethodNotAllowed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mock.MockBalanceService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewBalanceHandler(mockService)

			req := httptest.NewRequest(tt.method, "/api/user/withdrawals", nil)

			if tt.userID != nil {
				if uid, ok := tt.userID.(int64); ok {
					ctx := context.WithValue(req.Context(), "userID", uid)
					req = req.WithContext(ctx)
				}
			}

			w := httptest.NewRecorder()
			handler.GetWithdrawalsHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var withdrawals []model.Withdrawal
				err := json.Unmarshal(w.Body.Bytes(), &withdrawals)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedBody), len(withdrawals))
				for i, wd := range withdrawals {
					assert.Equal(t, tt.expectedBody[i].Order, wd.Order)
					assert.Equal(t, tt.expectedBody[i].Amount, wd.Amount)
					assert.WithinDuration(t, tt.expectedBody[i].ProcessedAt, wd.ProcessedAt, time.Second)
				}
			} else if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, w.Body.Bytes())
			} else if tt.expectedError != "" {
				var errResp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
				assert.Equal(t, tt.expectedError, errResp["error"])
			}
		})
	}
}
