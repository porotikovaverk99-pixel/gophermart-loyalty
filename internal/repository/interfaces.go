package repository

import (
	"context"
	"errors"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
)

var (
	ErrLoginAlreadyExists = errors.New("login dublicate")
	ErrUserNotFound       = errors.New("user not found")

	ErrNumberAlreadyExists = errors.New("number dublicate")
)

type UserRepository interface {
	CreateUser(ctx context.Context, login string, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, userID int64, number string) (int64, error)
	GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error)
	GetOrdersToProcess(ctx context.Context) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status string, accrual *float64) error
	UpdateLastChecked(ctx context.Context, orderID int64, time time.Time) error
	ScheduleNextCheck(ctx context.Context, orderID int64, nextCheck time.Time, retryCount int) error
}

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID int64) (float64, float64, error)
}
