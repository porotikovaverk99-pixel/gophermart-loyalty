package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
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

func (ps *BalancePostgresRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNum string, amount float64) error {

	tx, err := ps.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var existingWithdrawal bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM balance_transactions 
            WHERE order_number = $1 AND type = 'WITHDRAWAL'
        )`,
		orderNum).Scan(&existingWithdrawal)

	if err != nil {
		return fmt.Errorf("check existing withdrawal: %w", err)
	}

	if existingWithdrawal {
		return ErrOrderAlreadyWithdrawn
	}

	_, err = tx.Exec(ctx,
		`SELECT 1 FROM balance_transactions WHERE user_id = $1 FOR UPDATE`,
		userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("lock user transactions: %w", err)
	}

	var currentBalance float64
	err = tx.QueryRow(ctx,
		`SELECT 
            COALESCE(SUM(CASE WHEN type = 'ACCRUAL' THEN amount END), 0) -
            COALESCE(SUM(CASE WHEN type = 'WITHDRAWAL' THEN amount END), 0) as balance
         FROM balance_transactions 
         WHERE user_id = $1`,
		userID).Scan(&currentBalance)

	if err != nil {
		return fmt.Errorf("calculate balance: %w", err)
	}

	if currentBalance < amount {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO balance_transactions (user_id, type, order_number, amount)
         VALUES ($1, 'WITHDRAWAL', $2, $3)`,
		userID, orderNum, amount)

	if err != nil {
		return fmt.Errorf("create withdrawal transaction: %w", err)
	}

	return tx.Commit(ctx)
}

func (ps *BalancePostgresRepository) CreateAccrual(ctx context.Context, userID int64, orderNum string, amount float64) error {

	tx, err := ps.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM balance_transactions 
            WHERE order_number = $1 AND type = 'ACCRUAL'
        )`,
		orderNum).Scan(&exists)

	if err != nil {
		return fmt.Errorf("check duplicate accrual: %w", err)
	}

	if exists {
		return ErrAccrualAlreadyExists
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO balance_transactions (user_id, type, order_number, amount)
         VALUES ($1, 'ACCRUAL', $2, $3)`,
		userID, orderNum, amount)

	if err != nil {
		return fmt.Errorf("create accrual transaction: %w", err)
	}

	return tx.Commit(ctx)
}

func (ps *BalancePostgresRepository) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {

	rows, err := ps.pool.Query(ctx,
		`SELECT order_number, amount, processed_at
         FROM balance_transactions 
         WHERE user_id = $1 AND type = 'WITHDRAWAL'
		 ORDER BY processed_at DESC`,
		userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	var result []model.Withdrawal
	defer rows.Close()

	for rows.Next() {
		var withdrawal model.Withdrawal
		err := rows.Scan(
			&withdrawal.Order,
			&withdrawal.Amount,
			&withdrawal.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		result = append(result, withdrawal)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil

}
