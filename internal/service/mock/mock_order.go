package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
)

// MockBalanceRepo — мок с in-memory хранилищем.
type MockOrderRepo struct {
	mu     sync.RWMutex
	orders []model.Order
}

func NewMockOrderRepo() *MockOrderRepo {
	return &MockOrderRepo{
		orders: make([]model.Order, 0),
	}
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, userID int64, number string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tx := range m.orders {
		if tx.Number == number {
			return 0, repository.ErrNumberAlreadyExists
		}
	}

	id := int64(len(m.orders) + 1)

	m.orders = append(m.orders, model.Order{
		ID:          id,
		UserID:      userID,
		Number:      number,
		UploadedAt:  time.Now(),
		Status:      "NEW",
		NextCheckAt: nil,
		Accrual:     nil,
	})

	return id, nil
}

func (m *MockOrderRepo) GetOrderByNumber(ctx context.Context, number string) (model.Order, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tx := range m.orders {
		if tx.Number == number {
			return tx, nil
		}
	}

	return model.Order{}, repository.ErrOrderNotFound
}

func (m *MockOrderRepo) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []model.Order
	for _, tx := range m.orders {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}

	return result, nil
}

func (m *MockOrderRepo) GetOrdersToProcess(ctx context.Context) ([]model.Order, error) {
	return []model.Order{}, nil
}

func (m *MockOrderRepo) UpdateOrderStatus(ctx context.Context, id int64, status string, accrual *float64) error {
	return nil
}

func (m *MockOrderRepo) UpdateLastChecked(ctx context.Context, orderID int64, time time.Time) error {
	return nil
}

func (m *MockOrderRepo) ScheduleNextCheck(ctx context.Context, orderID int64, nextCheck time.Time, retryCount int) error {
	return nil
}

func (m *MockOrderRepo) MarkOrderAsFinal(ctx context.Context, orderID int64) error {
	return nil
}
