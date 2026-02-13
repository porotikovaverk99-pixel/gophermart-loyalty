package service

import (
	"context"
	"testing"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	mocks "github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		reqs      model.RequestAuth
		setupData func(*mocks.MockUserRepo)
		wantErr   error
	}{
		{
			name: "успешная регистрация",
			reqs: model.RequestAuth{Login: "test", Password: "123456"},
			setupData: func(m *mocks.MockUserRepo) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("123123"), bcrypt.DefaultCost)
				_, _ = m.CreateUser(ctx, "test123", string(hash))
			},
			wantErr: nil,
		},
		{
			name: "логин уже существует",
			reqs: model.RequestAuth{Login: "test123", Password: "123456"},
			setupData: func(m *mocks.MockUserRepo) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("123123"), bcrypt.DefaultCost)
				_, _ = m.CreateUser(ctx, "test123", string(hash))
			},
			wantErr: ErrLoginAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepo()
			tt.setupData(mockRepo)

			jwtManager := auth.NewJWTManager("", 30*time.Minute)
			service := NewAuthService(mockRepo, jwtManager)

			_, err := service.Register(ctx, tt.reqs)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "should return correct error")
			} else {
				assert.NoError(t, err, "should not return error")
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		reqs      model.RequestAuth
		setupData func(*mocks.MockUserRepo)
		wantErr   error
	}{
		{
			name: "успешная аутентификация",
			reqs: model.RequestAuth{Login: "test123", Password: "123123"},
			setupData: func(m *mocks.MockUserRepo) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("123123"), bcrypt.DefaultCost)
				_, _ = m.CreateUser(ctx, "test123", string(hash))
			},
			wantErr: nil,
		},
		{
			name: "пользователь не найден",
			reqs: model.RequestAuth{Login: "test124", Password: "123456"},
			setupData: func(m *mocks.MockUserRepo) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("123123"), bcrypt.DefaultCost)
				_, _ = m.CreateUser(ctx, "test123", string(hash))
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "неверный пароль",
			reqs: model.RequestAuth{Login: "test123", Password: "123456"},
			setupData: func(m *mocks.MockUserRepo) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("123123"), bcrypt.DefaultCost)
				_, _ = m.CreateUser(ctx, "test123", string(hash))
			},
			wantErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepo()
			tt.setupData(mockRepo)

			jwtManager := auth.NewJWTManager("", 30*time.Minute)
			service := NewAuthService(mockRepo, jwtManager)

			_, err := service.Login(ctx, tt.reqs)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "should return correct error")
			} else {
				assert.NoError(t, err, "should not return error")
			}
		})
	}
}
