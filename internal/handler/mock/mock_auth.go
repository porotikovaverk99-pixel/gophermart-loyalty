package mock

import (
	"context"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
)

type MockAuthService struct {
	ShouldFail bool
	FailWith   error
	Token      string
}

func (m *MockAuthService) Register(ctx context.Context, reqs model.RequestAuth) (string, error) {
	if m.ShouldFail {
		return "", m.FailWith
	}
	return m.Token, nil
}

func (m *MockAuthService) Login(ctx context.Context, reqs model.RequestAuth) (string, error) {
	if m.ShouldFail {
		return "", m.FailWith
	}
	return m.Token, nil
}
