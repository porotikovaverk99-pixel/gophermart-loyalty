// Package service реализует бизнес-логику системы лояльности.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/validator"
)

var (
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrOrderAlreadyWithdrawn = errors.New("order already withdrawn")
	ErrInvalidAmount         = errors.New("amount must be positive")
	ErrAccrualAlreadyExists  = errors.New("accrual already exists for order")
)

// BalanceService управляет балансом пользователей.
type BalanceService struct {
	repo repository.BalanceRepository
}

// NewBalanceService создает новый сервис баланса.
func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

// GetUserBalance возвращает текущий баланс и сумму списаний пользователя.
func (s *BalanceService) GetUserBalance(ctx context.Context, userID int64) (model.BalanceResponse, error) {
	current, withdrawn, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return model.BalanceResponse{}, fmt.Errorf("get balance: %w", err)
	}
	return model.BalanceResponse{Current: current, Withdrawn: withdrawn}, nil
}

// CreateWithdraw списывает баллы с баланса пользователя.
// Ошибки: ErrInsufficientFunds, ErrOrderAlreadyWithdrawn.
func (s *BalanceService) CreateWithdraw(ctx context.Context, reqs model.WithdrawRequest, userID int64) error {

	if !validator.Luhn(reqs.Order) {
		return ErrInvalidOrderNumber
	}

	if reqs.Sum <= 0 {
		return ErrInvalidAmount
	}

	err := s.repo.CreateWithdrawal(ctx, userID, reqs.Order, reqs.Sum)
	if err != nil {

		switch {
		case errors.Is(err, repository.ErrInsufficientFunds):
			return ErrInsufficientFunds
		case errors.Is(err, repository.ErrOrderAlreadyWithdrawn):
			return ErrOrderAlreadyWithdrawn
		default:
			return fmt.Errorf("withdrawal failed: %w", err)
		}
	}
	return nil
}

// CreateAccrual начисляет баллы пользователю за обработанный заказ.
func (s *BalanceService) CreateAccrual(ctx context.Context, userID int64, orderNum string, amount float64) error {

	err := s.repo.CreateAccrual(ctx, userID, orderNum, amount)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAccrualAlreadyExists):
			return ErrAccrualAlreadyExists
		default:
			return fmt.Errorf("create accrual: %w", err)
		}
	}

	return nil
}

// GetUserWithdrawals возвращает все списания пользователя.
func (s *BalanceService) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {

	result, err := s.repo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}

	return result, nil
}
