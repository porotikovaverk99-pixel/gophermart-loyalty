package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BalancePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) *BalancePostgresRepository {
	return &BalancePostgresRepository{pool: pool}
}

func (ps *BalancePostgresRepository) GetUserBalance(ctx context.Context, userID int64) (float64, float64, error) {

	var currentBalance, withdrawn float64

	err := ps.pool.QueryRow(ctx,
		`SELECT 
            COALESCE(SUM(CASE WHEN type = 'ACCRUAL' THEN amount ELSE 0 END), 0) as current,
            COALESCE(SUM(CASE WHEN type = 'WITHDRAWAL' THEN amount ELSE 0 END), 0) as withdrawn
         FROM balance_transactions 
         WHERE user_id = $1`,
		userID).Scan(&currentBalance, &withdrawn)

	if err != nil {
		return 0, 0, fmt.Errorf("calculate balance: %w", err)
	}

	finalBalance := currentBalance - withdrawn
	return finalBalance, withdrawn, nil

}
