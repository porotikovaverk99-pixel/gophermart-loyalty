package repository

import (
	"context"
	"errors"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{pool: pool}
}

func (ps *UserPostgresRepository) CreateUser(ctx context.Context, login string, passwordHash string) (int64, error) {
	var id int64

	err := ps.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) 
         VALUES ($1, $2)
         ON CONFLICT (login) DO NOTHING
         RETURNING id`,
		login, passwordHash).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrLoginAlreadyExists
		}
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

func (ps *UserPostgresRepository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User

	err := ps.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1`,
		login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
