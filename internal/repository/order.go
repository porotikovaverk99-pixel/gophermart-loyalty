package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderPostgresRepository {
	return &OrderPostgresRepository{pool: pool}
}

func (ps *OrderPostgresRepository) CreateOrder(ctx context.Context, userID int64, number string) (int64, error) {
	var id int64

	err := ps.pool.QueryRow(ctx,
		`INSERT INTO orders (user_id, number) 
         VALUES ($1, $2)
         ON CONFLICT (number) DO NOTHING
         RETURNING id`,
		userID, number).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNumberAlreadyExists
		}
		return 0, fmt.Errorf("failed to save order: %w", err)
	}

	return id, nil
}

func (ps *OrderPostgresRepository) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {

	rows, err := ps.pool.Query(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at, last_checked_at, next_check_at, retry_count
		 FROM orders
         WHERE user_id = $1
		 ORDER BY uploaded_at DESC`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var result []model.Order
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.LastCheckedAt,
			&order.NextCheckAt,
			&order.RetryCount)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		result = append(result, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil

}

func (ps *OrderPostgresRepository) GetOrdersToProcess(ctx context.Context) ([]model.Order, error) {

	rows, err := ps.pool.Query(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at, last_checked_at, next_check_at, retry_count
		 FROM orders
         WHERE status IN ('NEW', 'PROCESSING') 
		 AND next_check_at <= $1
		 FOR UPDATE SKIP LOCKED`, time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	var result []model.Order
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
			&order.LastCheckedAt,
			&order.NextCheckAt,
			&order.RetryCount)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		result = append(result, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil

}

func (ps *OrderPostgresRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status string, accrual *float64) error {
	_, err := ps.pool.Exec(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE id = $3",
		status, accrual, orderID)
	return err
}

func (ps *OrderPostgresRepository) UpdateLastChecked(ctx context.Context, orderID int64, time time.Time) error {
	_, err := ps.pool.Exec(ctx,
		"UPDATE orders SET last_checked_at = $1 WHERE id = $2",
		time, orderID)
	return err
}

func (ps *OrderPostgresRepository) ScheduleNextCheck(ctx context.Context, orderID int64, nextCheck time.Time, retryCount int) error {
	_, err := ps.pool.Exec(ctx,
		"UPDATE orders SET next_check_at = $1, retry_count = $2 WHERE id = $3",
		nextCheck, retryCount, orderID)
	return err
}
