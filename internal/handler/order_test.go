package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/handler/mock"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_BaseOrderHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    []byte
		userID         int64
		setupMock      func(*mock.MockOrderService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "успешная загрузка нового заказа",
			method:      http.MethodPost,
			requestBody: []byte("4111111111111111"),
			userID:      1,
			setupMock: func(m *mock.MockOrderService) {
				m.UploadOrderResult = 1
				m.UploadOrderError = nil
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "заказ уже загружен этим пользователем",
			method:      http.MethodPost,
			requestBody: []byte("4111111111111111"),
			userID:      1,
			setupMock: func(m *mock.MockOrderService) {
				m.UploadOrderError = service.ErrNumberAlreadyExists
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "заказ принадлежит другому пользователю",
			method:      http.MethodPost,
			requestBody: []byte("4111111111111111"),
			userID:      1,
			setupMock: func(m *mock.MockOrderService) {
				m.UploadOrderError = service.ErrOrderBelongsToAnother
			},
			expectedStatus: http.StatusConflict,
			expectedError:  service.ErrOrderBelongsToAnother.Error(),
		},
		{
			name:        "невалидный номер заказа",
			method:      http.MethodPost,
			requestBody: []byte("123"),
			userID:      1,
			setupMock: func(m *mock.MockOrderService) {
				m.UploadOrderError = service.ErrInvalidOrderNumber
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  service.ErrInvalidOrderNumber.Error(),
		},
		{
			name:           "пустое тело запроса",
			method:         http.MethodPost,
			requestBody:    []byte(""),
			userID:         1,
			setupMock:      func(m *mock.MockOrderService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  http.StatusText(http.StatusBadRequest),
		},
		{
			name:        "метод (GET)",
			method:      http.MethodGet,
			requestBody: nil,
			userID:      1,
			setupMock: func(m *mock.MockOrderService) {
				m.GetUserOrdersResult = []model.Order{
					{
						Number:     "4111111111111111",
						Status:     "PROCESSED",
						Accrual:    float64Ptr(500),
						UploadedAt: time.Now(),
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "GET пустой список заказов",
			method:      http.MethodGet,
			requestBody: nil,
			userID:      1,
			setupMock: func(m *mock.MockOrderService) {
				m.GetUserOrdersResult = []model.Order{}
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mock.MockOrderService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewOrderHandler(mockService)

			var req *http.Request
			if tt.requestBody != nil {
				req = httptest.NewRequest(tt.method, "/api/user/orders", bytes.NewReader(tt.requestBody))
				req.Header.Set("Content-Type", "text/plain")
			} else {
				req = httptest.NewRequest(tt.method, "/api/user/orders", nil)
			}

			ctx := context.WithValue(req.Context(), "userID", tt.userID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.BaseOrderHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK &&
				tt.expectedStatus != http.StatusAccepted &&
				tt.expectedStatus != http.StatusNoContent {
				var errResp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, errResp["error"])
				}
			}

			if tt.method == http.MethodGet && tt.expectedStatus == http.StatusOK {
				var orders []model.Order
				err := json.Unmarshal(w.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.Equal(t, len(mockService.GetUserOrdersResult), len(orders))
			}
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}
