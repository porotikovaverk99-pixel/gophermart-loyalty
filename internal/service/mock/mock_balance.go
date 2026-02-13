package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
)

// MockBalanceRepo — мок с in-memory хранилищем.
type MockBalanceRepo struct {
	mu           sync.RWMutex
	transactions []model.BalanceTransaction
	users        map[int64]*userBalance
}

type userBalance struct {
	accruals    float64
	withdrawals float64
}

func NewMockBalanceRepo() *MockBalanceRepo {
	return &MockBalanceRepo{
		transactions: make([]model.BalanceTransaction, 0),
		users:        make(map[int64]*userBalance),
	}
}

func (m *MockBalanceRepo) GetUserBalance(ctx context.Context, userID int64) (float64, float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var accruals, withdrawals float64
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			if tx.Type == "ACCRUAL" {
				accruals += tx.Amount
			} else if tx.Type == "WITHDRAWAL" {
				withdrawals += tx.Amount
			}
		}
	}

	current := accruals - withdrawals
	return current, withdrawals, nil
}

func (m *MockBalanceRepo) CreateWithdrawal(ctx context.Context, userID int64, orderNum string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tx := range m.transactions {
		if tx.OrderNumber == orderNum && tx.Type == "WITHDRAWAL" {
			return repository.ErrOrderAlreadyWithdrawn
		}
	}

	current, _, _ := m.calculateBalance(userID)
	if current < amount {
		return repository.ErrInsufficientFunds
	}

	m.transactions = append(m.transactions, model.BalanceTransaction{
		ID:          int64(len(m.transactions) + 1),
		UserID:      userID,
		Type:        "WITHDRAWAL",
		OrderNumber: orderNum,
		Amount:      amount,
		ProcessedAt: time.Now(),
	})

	return nil
}

func (m *MockBalanceRepo) CreateAccrual(ctx context.Context, userID int64, orderNum string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tx := range m.transactions {
		if tx.OrderNumber == orderNum && tx.Type == "ACCRUAL" {
			return repository.ErrAccrualAlreadyExists
		}
	}

	m.transactions = append(m.transactions, model.BalanceTransaction{
		ID:          int64(len(m.transactions) + 1),
		UserID:      userID,
		Type:        "ACCRUAL",
		OrderNumber: orderNum,
		Amount:      amount,
		ProcessedAt: time.Now(),
	})

	return nil
}

func (m *MockBalanceRepo) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []model.Withdrawal
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			if tx.Type == "WITHDRAWAL" {
				var withdrawal model.Withdrawal
				withdrawal.Order = tx.OrderNumber
				withdrawal.Amount = tx.Amount
				withdrawal.ProcessedAt = tx.ProcessedAt
				result = append(result, withdrawal)
			}
		}
	}

	return result, nil
}

func (m *MockBalanceRepo) calculateBalance(userID int64) (float64, float64, float64) {
	var accruals, withdrawals float64
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			if tx.Type == "ACCRUAL" {
				accruals += tx.Amount
			} else if tx.Type == "WITHDRAWAL" {
				withdrawals += tx.Amount
			}
		}
	}
	return accruals - withdrawals, accruals, withdrawals
}
