package service

import (
	"context"
	"fmt"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
)

var ()

type BalanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID int64) (model.BalanceResponse, error) {
	current, withdrawn, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return model.BalanceResponse{}, fmt.Errorf("get balance: %w", err)
	}
	return model.BalanceResponse{Current: current, Withdrawn: withdrawn}, nil
}

func (s *BalanceService) Withdraw(ctx context.Context, reqs model.ResponseWithdraw, userID int64) error {
	return nil
}
