package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(dsn string) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Database{pool: pool}, nil
}

func (db *Database) GetPool() *pgxpool.Pool {
	return db.pool
}

func (db *Database) Close() {
	db.pool.Close()
}

func (db *Database) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
