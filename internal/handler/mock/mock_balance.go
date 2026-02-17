package mock

import (
	"context"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
)

type MockBalanceService struct {
	GetBalanceResult model.BalanceResponse
	GetBalanceError  error

	CreateWithdrawError error

	GetUserWithdrawalsResult []model.Withdrawal
	GetUserWithdrawalsError  error
}

func (m *MockBalanceService) GetUserBalance(ctx context.Context, userID int64) (model.BalanceResponse, error) {
	return m.GetBalanceResult, m.GetBalanceError
}

func (m *MockBalanceService) CreateWithdraw(ctx context.Context, reqs model.WithdrawRequest, userID int64) error {
	return m.CreateWithdrawError
}

func (m *MockBalanceService) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	return m.GetUserWithdrawalsResult, m.GetUserWithdrawalsError
}
