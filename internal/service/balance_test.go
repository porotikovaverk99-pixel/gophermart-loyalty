package service

import (
	"context"
	"testing"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	mocks "github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service/mock"
	"github.com/stretchr/testify/assert"
)

func TestBalanceService_GetUserBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		setupData func(*mocks.MockBalanceRepo)
		want      model.BalanceResponse
		wantErr   bool
	}{
		{
			name:   "начисления и списания",
			userID: 1,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 1000)
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
				_ = m.CreateWithdrawal(ctx, 1, "378282246310005", 300)
			},
			want: model.BalanceResponse{
				Current:   1200,
				Withdrawn: 300,
			},
			wantErr: false,
		},
		{
			name:   "пользователь без операций",
			userID: 1,
			setupData: func(m *mocks.MockBalanceRepo) {
			},
			want: model.BalanceResponse{
				Current:   0,
				Withdrawn: 0,
			},
			wantErr: false,
		},
		{
			name:   "только начисления",
			userID: 1,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4111111111111111", 2000)
			},
			want: model.BalanceResponse{
				Current:   2000,
				Withdrawn: 0,
			},
			wantErr: false,
		},
		{
			name:   "пустой остаток",
			userID: 1,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
				_ = m.CreateWithdrawal(ctx, 1, "4111111111111111", 500)
			},
			want: model.BalanceResponse{
				Current:   0,
				Withdrawn: 500,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockBalanceRepo()
			tt.setupData(mockRepo)
			service := NewBalanceService(mockRepo)

			got, err := service.GetUserBalance(ctx, tt.userID)

			assert.Equal(t, tt.wantErr, err != nil, "error presence mismatch")

			if err == nil {
				assert.Equal(t, tt.want.Current, got.Current, "current balance mismatch")
				assert.Equal(t, tt.want.Withdrawn, got.Withdrawn, "withdrawn balance mismatch")
			}
		})
	}
}

func TestBalanceService_CreateAccrual(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      int64
		orderNumber string
		sum         float64
		setupData   func(*mocks.MockBalanceRepo)
		wantErr     error
	}{
		{
			name:        "начисления и списания",
			userID:      1,
			orderNumber: "49927398716",
			sum:         100,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 1000)
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
				_ = m.CreateWithdrawal(ctx, 1, "378282246310005", 300)
			},
			wantErr: nil,
		},
		{
			name:        "пользователь без операций",
			userID:      1,
			orderNumber: "4111111111111111",
			sum:         100,
			setupData: func(m *mocks.MockBalanceRepo) {
			},
			wantErr: nil,
		},
		{
			name:        "только начисления",
			userID:      1,
			orderNumber: "5555555555554444",
			sum:         100,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 2000)
			},
			wantErr: nil,
		},
		{
			name:        "повторное начисление",
			userID:      1,
			orderNumber: "378282246310005",
			sum:         100,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 500)
				_ = m.CreateAccrual(ctx, 1, "378282246310005", 500)
			},
			wantErr: ErrAccrualAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockBalanceRepo()
			tt.setupData(mockRepo)
			service := NewBalanceService(mockRepo)

			err := service.CreateAccrual(ctx, tt.userID, tt.orderNumber, tt.sum)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "should return correct error")
			} else {
				assert.NoError(t, err, "should not return error")
			}
		})
	}
}

func TestBalanceService_CreateWithdraw(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		reqs      model.WithdrawRequest
		setupData func(*mocks.MockBalanceRepo)
		wantErr   error
	}{
		{
			name:   "начисления и списания",
			userID: 1,
			reqs:   model.WithdrawRequest{Order: "3530111333300000", Sum: 100},
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 1000)
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
				_ = m.CreateWithdrawal(ctx, 1, "378282246310005", 300)
			},
			wantErr: nil,
		},
		{
			name:   "повторное списание",
			userID: 1,
			reqs:   model.WithdrawRequest{Order: "378282246310005", Sum: 100},
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 1000)
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
				_ = m.CreateWithdrawal(ctx, 1, "378282246310005", 300)
			},
			wantErr: ErrOrderAlreadyWithdrawn,
		},
		{
			name:   "недостаточно средств",
			userID: 1,
			reqs:   model.WithdrawRequest{Order: "49927398716", Sum: 3000},
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 2000)
			},
			wantErr: ErrInsufficientFunds,
		},
		{
			name:   "невалидная сумма",
			userID: 1,
			reqs:   model.WithdrawRequest{Order: "4111111111111111", Sum: -100},
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 500)
				_ = m.CreateAccrual(ctx, 1, "378282246310005", 500)
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name:   "невалидный номер заказа",
			userID: 1,
			reqs:   model.WithdrawRequest{Order: "4111111111111112", Sum: 100},
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 500)
				_ = m.CreateAccrual(ctx, 1, "378282246310005", 500)
			},
			wantErr: ErrInvalidOrderNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockBalanceRepo()
			tt.setupData(mockRepo)
			service := NewBalanceService(mockRepo)

			err := service.CreateWithdraw(ctx, tt.reqs, tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "should return correct error")
			} else {
				assert.NoError(t, err, "should not return error")
			}
		})
	}
}

func TestBalanceService_GetUserWithdrawals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		setupData func(*mocks.MockBalanceRepo)
		want      []model.Withdrawal
		wantErr   bool
	}{
		{
			name:   "начисления и списания",
			userID: 1,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 1000)
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
				_ = m.CreateWithdrawal(ctx, 1, "378282246310005", 300)
			},
			want: []model.Withdrawal{{
				Order:  "378282246310005",
				Amount: 300,
			}},
			wantErr: false,
		},
		{
			name:   "нет списаний",
			userID: 1,
			setupData: func(m *mocks.MockBalanceRepo) {
				_ = m.CreateAccrual(ctx, 1, "4561261212345467", 1000)
				_ = m.CreateAccrual(ctx, 1, "5555555555554444", 500)
			},
			want:    []model.Withdrawal{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockBalanceRepo()
			tt.setupData(mockRepo)
			service := NewBalanceService(mockRepo)

			got, err := service.GetUserWithdrawals(ctx, tt.userID)

			assert.Equal(t, tt.wantErr, err != nil, "error presence mismatch")

			if err == nil {

				assert.Len(t, got, len(tt.want), "number of withdrawals mismatch")

				for i, w := range tt.want {

					assert.Equal(t, w.Order, got[i].Order, "order mismatch")
					assert.Equal(t, w.Amount, got[i].Amount, "amount mismatch")

					assert.NotZero(t, got[i].ProcessedAt, "processed_at should not be zero")

					assert.WithinDuration(t, time.Now(), got[i].ProcessedAt, 5*time.Second)
				}
			}
		})
	}
}
