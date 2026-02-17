package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
)

// MockUserRepo — мок с in-memory хранилищем.
type MockUserRepo struct {
	mu    sync.RWMutex
	users []model.User
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make([]model.User, 0),
	}
}

func (m *MockUserRepo) CreateUser(ctx context.Context, login string, passwordHash string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tx := range m.users {
		if tx.Login == login {
			return 0, repository.ErrLoginAlreadyExists
		}
	}

	id := int64(len(m.users) + 1)

	m.users = append(m.users, model.User{
		ID:           id,
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})

	return id, nil
}

func (m *MockUserRepo) GetUserByLogin(ctx context.Context, login string) (model.User, error) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tx := range m.users {
		if tx.Login == login {
			return tx, nil
		}
	}

	return model.User{}, repository.ErrUserNotFound
}
