package service

import (
	"context"
	"testing"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/client"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	mocks "github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service/mock"
	"github.com/stretchr/testify/assert"
)

func TestOrderService_UploadOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		number    string
		setupData func(*mocks.MockOrderRepo)
		want      int64
		wantErr   error
	}{
		{
			name:   "новый заказ",
			userID: 1,
			number: "5555555555554444",
			setupData: func(m *mocks.MockOrderRepo) {
				_, _ = m.CreateOrder(ctx, 1, "4111111111111111")
			},
			want:    2,
			wantErr: nil,
		},
		{
			name:   "заказ уже загружен",
			userID: 1,
			number: "4111111111111111",
			setupData: func(m *mocks.MockOrderRepo) {
				_, _ = m.CreateOrder(ctx, 1, "4111111111111111")
			},
			want:    0,
			wantErr: ErrNumberAlreadyExists,
		},
		{
			name:   "заказ уже загружен другим пользователем",
			userID: 2,
			number: "4111111111111111",
			setupData: func(m *mocks.MockOrderRepo) {
				_, _ = m.CreateOrder(ctx, 1, "4111111111111111")
			},
			want:    0,
			wantErr: ErrOrderBelongsToAnother,
		},
		{
			name:   "невалидный номер заказа",
			userID: 2,
			number: "4111111111111112",
			setupData: func(m *mocks.MockOrderRepo) {
				_, _ = m.CreateOrder(ctx, 1, "4111111111111111")
			},
			want:    0,
			wantErr: ErrInvalidOrderNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockOrderRepo()
			tt.setupData(mockRepo)

			mockBalanceRepo := mocks.NewMockBalanceRepo()
			balanceService := NewBalanceService(mockBalanceRepo)

			accrualClient := client.NewAccrualClient("http://localhost:8081")
			service := NewOrderService(mockRepo, accrualClient, balanceService, nil, 100, 5, 5)

			got, err := service.UploadOrder(ctx, tt.userID, tt.number)

			assert.Equal(t, tt.want, got, "id mismatch")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "should return correct error")
			} else {
				assert.NoError(t, err, "should not return error")
			}

		})
	}
}

func TestOrderService_GetUserOrders(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		setupData func(*mocks.MockOrderRepo)
		want      []model.Order
		wantErr   bool
	}{
		{
			name:   "есть заказы",
			userID: 1,
			setupData: func(m *mocks.MockOrderRepo) {
				_, _ = m.CreateOrder(ctx, 1, "4111111111111111")
				_, _ = m.CreateOrder(ctx, 1, "5555555555554444")
			},
			want: []model.Order{
				{
					ID:          1,
					UserID:      1,
					Number:      "4111111111111111",
					Status:      "NEW",
					NextCheckAt: nil,
					Accrual:     nil,
				},
				{
					ID:          2,
					UserID:      1,
					Number:      "5555555555554444",
					Status:      "NEW",
					NextCheckAt: nil,
					Accrual:     nil,
				},
			},
			wantErr: false,
		},
		{
			name:   "нет заказов",
			userID: 2,
			setupData: func(m *mocks.MockOrderRepo) {
				_, _ = m.CreateOrder(ctx, 1, "4111111111111111")
			},
			want:    []model.Order{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockOrderRepo()
			tt.setupData(mockRepo)

			mockBalanceRepo := mocks.NewMockBalanceRepo()
			balanceService := NewBalanceService(mockBalanceRepo)

			accrualClient := client.NewAccrualClient("http://localhost:8081")
			service := NewOrderService(mockRepo, accrualClient, balanceService, nil, 100, 5, 5)

			got, err := service.GetUserOrders(ctx, tt.userID)

			assert.Equal(t, tt.wantErr, err != nil, "error presence mismatch")

			if err == nil {

				assert.Len(t, got, len(tt.want), "number of orders mismatch")

				for i, w := range tt.want {

					assert.Equal(t, w.ID, got[i].ID, "id mismatch")
					assert.Equal(t, w.Number, got[i].Number, "number mismatch")
					assert.Equal(t, w.Status, got[i].Status, "status mismatch")

					assert.NotZero(t, got[i].UploadedAt, "processed_at should not be zero")

					assert.WithinDuration(t, time.Now(), got[i].UploadedAt, 5*time.Second)
				}
			}

		})
	}
}
