package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(dsn string) (*Database, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	db := &Database{pool: pool}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := db.runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
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

func (db *Database) runMigrations(dsn string) error {
	needMigrations, err := db.checkIfMigrationsNeeded()
	if err != nil {
		return fmt.Errorf("failed to check migrations status: %w", err)
	}

	if !needMigrations {
		return nil
	}

	migrationsPath, err := findMigrationsPath()
	if err != nil {
		return fmt.Errorf("failed to find migrations directory: %w", err)
	}

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (db *Database) checkIfMigrationsNeeded() (bool, error) {
	ctx := context.Background()

	var tableExists bool
	err := db.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_migrations'
		)
	`).Scan(&tableExists)
	if err != nil {
		return false, fmt.Errorf("check schema_migrations table: %w", err)
	}

	if !tableExists {
		return true, nil
	}

	var dirty bool
	err = db.pool.QueryRow(ctx, `SELECT dirty FROM schema_migrations`).Scan(&dirty)
	if err != nil {
		return false, fmt.Errorf("check migrations dirty state: %w", err)
	}

	if dirty {
		return false, fmt.Errorf("migrations are in dirty state")
	}

	return false, nil
}

func findMigrationsPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	migrationsPath := filepath.Join(exeDir, "migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "migrations"
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			return "", fmt.Errorf("migrations directory not found")
		}
	}

	return migrationsPath, nil
}