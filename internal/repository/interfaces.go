// Package repository определяет интерфейсы для работы с БД.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
)

var (
	// Ошибки пользователей
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrUserNotFound       = errors.New("user not found")

	// Ошибки заказов
	ErrNumberAlreadyExists = errors.New("order number already exists")
	ErrOrderNotFound       = errors.New("order not found")

	// Ошибки баланса
	ErrOrderAlreadyWithdrawn = errors.New("order already withdrawn")
	ErrAccrualAlreadyExists  = errors.New("accrual already exists for order")
	ErrInsufficientFunds     = errors.New("insufficient funds")
)

// UserRepository — операции с пользователями.
type UserRepository interface {
	// CreateUser создает нового пользователя.
	CreateUser(ctx context.Context, login string, passwordHash string) (int64, error)

	// GetUserByLogin возвращает пользователя по логину.
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
}

// OrderRepository — операции с заказами.
type OrderRepository interface {
	// CreateOrder создает новый заказ.
	CreateOrder(ctx context.Context, userID int64, number string) (int64, error)

	// GetOrderByNumber возвращает заказ по номеру.
	GetOrderByNumber(ctx context.Context, number string) (model.Order, error)

	// GetUserOrders возвращает заказы пользователя.
	GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error)

	// GetOrdersToProcess возвращает заказы, готовые к проверке.
	GetOrdersToProcess(ctx context.Context) ([]model.Order, error)

	// UpdateOrderStatus обновляет статус и начисление.
	UpdateOrderStatus(ctx context.Context, id int64, status string, accrual *float64) error

	// UpdateLastChecked обновляет время последней проверки.
	UpdateLastChecked(ctx context.Context, orderID int64, time time.Time) error

	// ScheduleNextCheck планирует следующую проверку.
	ScheduleNextCheck(ctx context.Context, orderID int64, nextCheck time.Time, retryCount int) error

	// MarkOrderAsFinal фиксирует заказ как обработанный.
	MarkOrderAsFinal(ctx context.Context, orderID int64) error
}

// BalanceRepository — операции с балансом.
type BalanceRepository interface {
	// GetUserBalance возвращает текущий баланс и сумму списаний.
	GetUserBalance(ctx context.Context, userID int64) (float64, float64, error)

	// CreateAccrual начисляет баллы за заказ.
	CreateAccrual(ctx context.Context, userID int64, orderNum string, amount float64) error

	// CreateWithdrawal списывает баллы.
	CreateWithdrawal(ctx context.Context, userID int64, orderNum string, amount float64) error

	// GetUserWithdrawals возвращает списания пользователя.
	GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
