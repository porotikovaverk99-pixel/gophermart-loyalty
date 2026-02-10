package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidLogin       = errors.New("invalid login")
	ErrInvalidPassword    = errors.New("invalid password")
)

type AuthService struct {
	repo    repository.UserRepository
	manager auth.Manager
}

func NewAuthService(repo repository.UserRepository, manager auth.Manager) *AuthService {
	return &AuthService{
		repo:    repo,
		manager: manager,
	}
}

func (s *AuthService) Register(ctx context.Context, reqs model.RequestAuth) (string, error) {

	if err := validateAuthRequest(reqs); err != nil {
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

func validateAuthRequest(reqs model.RequestAuth) error {
	if reqs.Login == "" || reqs.Password == "" {
		return ErrInvalidInput
	}

	if len(reqs.Login) < 3 || len(reqs.Login) > 50 {
		return ErrInvalidLogin
	}

	if len(reqs.Password) < 6 {
		return ErrInvalidPassword
	}

	return nil
}
