package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLuhn(t *testing.T) {

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "валидный номер Visa",
			number: "4111111111111111",
			want:   true,
		},
		{
			name:   "валидный номер MasterCard",
			number: "5555555555554444",
			want:   true,
		},
		{
			name:   "валидный номер American Express",
			number: "378282246310005",
			want:   true,
		},
		{
			name:   "валидный короткий номер",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "невалидный номер",
			number: "1234567890123456",
			want:   false,
		},
		{
			name:   "пустая строка",
			number: "",
			want:   false,
		},
		{
			name:   "номер с буквами",
			number: "4111a11111111111",
			want:   false,
		},
		{
			name:   "номер с пробелами",
			number: "4111 1111 1111 1111",
			want:   false,
		},
		{
			name:   "слишком короткий номер",
			number: "123",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := Luhn(tt.number)

			assert.Equal(t, tt.want, got)
		})
	}
}
