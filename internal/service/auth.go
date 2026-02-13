// Package service реализует бизнес-логику системы лояльности.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidLogin       = errors.New("invalid login")
	ErrInvalidPassword    = errors.New("invalid password")
)

// AuthService отвечает за регистрацию и аутентификацию пользователей.
type AuthService struct {
	repo    repository.UserRepository
	manager auth.Manager
}

// NewAuthService создает новый сервис аутентификации.
func NewAuthService(repo repository.UserRepository, manager auth.Manager) *AuthService {
	return &AuthService{
		repo:    repo,
		manager: manager,
	}
}

// GetManager возвращает менеджер аутентификации.
func (s *AuthService) GetManager() auth.Manager {
	return s.manager
}

// Register регистрирует нового пользователя.
// Возвращает JWT-токен при успехе.
// Ошибки: ErrInvalidInput, ErrInvalidLogin, ErrInvalidPassword, ErrLoginAlreadyExists.
func (s *AuthService) Register(ctx context.Context, reqs model.RequestAuth) (string, error) {

	if err := validator.ValidateAuth(reqs); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(reqs.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	userID, err := s.repo.CreateUser(ctx, reqs.Login, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExists) {
			return "", ErrLoginAlreadyExists
		}
		return "", fmt.Errorf("create user: %w", err)
	}
	token, err := s.manager.Generate(userID, reqs.Login)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

// Login аутентифицирует пользователя.
// Возвращает JWT-токен при успехе.
// Ошибки: ErrInvalidInput, ErrInvalidCredentials.
func (s *AuthService) Login(ctx context.Context, reqs model.RequestAuth) (string, error) {

	user, err := s.repo.GetUserByLogin(ctx, reqs.Login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(reqs.Password))

	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.manager.Generate(user.ID, reqs.Login)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}
