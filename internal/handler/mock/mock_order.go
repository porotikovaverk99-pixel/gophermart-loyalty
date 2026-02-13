package mock

import (
	"context"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
)

type MockOrderService struct {
	UploadOrderResult int64
	UploadOrderError  error

	GetUserOrdersResult []model.Order
	GetUserOrdersError  error
}

func (m *MockOrderService) UploadOrder(ctx context.Context, userID int64, number string) (int64, error) {
	return m.UploadOrderResult, m.UploadOrderError
}

func (m *MockOrderService) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return m.GetUserOrdersResult, m.GetUserOrdersError
}
