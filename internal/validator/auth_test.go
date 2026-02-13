package validator

import (
	"testing"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestValidateAuth(t *testing.T) {

	tests := []struct {
		name    string
		reqs    model.RequestAuth
		wantErr error
	}{
		{
			name:    "пустые данные",
			reqs:    model.RequestAuth{Login: "", Password: ""},
			wantErr: ErrInvalidInput,
		},
		{
			name:    "логин короче 3 символов",
			reqs:    model.RequestAuth{Login: "12", Password: "123456"},
			wantErr: ErrInvalidLogin,
		},
		{
			name:    "логин длиннее 50 символов",
			reqs:    model.RequestAuth{Login: "123456789012345678901234567890123456789012345678901", Password: "123456"},
			wantErr: ErrInvalidLogin,
		},
		{
			name:    "пароль короче 6 символов",
			reqs:    model.RequestAuth{Login: "test", Password: "12345"},
			wantErr: ErrInvalidPassword,
		},
		{
			name:    "валидные данные",
			reqs:    model.RequestAuth{Login: "test", Password: "123456"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := ValidateAuth(tt.reqs)

			if tt.wantErr != nil {
				assert.ErrorIs(t, got, tt.wantErr, "should return correct error")
			} else {
				assert.NoError(t, got, "should not return error")
			}
		})
	}
}
