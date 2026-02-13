package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/handler/mock"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_RegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    []byte
		setupMock      func(*mock.MockAuthService)
		expectedStatus int
		expectedHeader string
		expectedError  string
	}{
		{
			name:   "успешная регистрация",
			method: http.MethodPost,
			requestBody: func() []byte {
				b, _ := json.Marshal(model.RequestAuth{
					Login:    "testuser",
					Password: "password123",
				})
				return b
			}(),
			setupMock: func(m *mock.MockAuthService) {
				m.Token = "valid.jwt.token"
				m.ShouldFail = false
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "Bearer valid.jwt.token",
		},
		{
			name:           "неверный формат JSON",
			method:         http.MethodPost,
			requestBody:    []byte(`{"login": "test",`),
			setupMock:      func(m *mock.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
			expectedError:  http.StatusText(http.StatusBadRequest),
		},
		{
			name:   "логин уже существует",
			method: http.MethodPost,
			requestBody: func() []byte {
				b, _ := json.Marshal(model.RequestAuth{
					Login:    "existing",
					Password: "pass123",
				})
				return b
			}(),
			setupMock: func(m *mock.MockAuthService) {
				m.ShouldFail = true
				m.FailWith = service.ErrLoginAlreadyExists
			},
			expectedStatus: http.StatusConflict,
			expectedHeader: "",
			expectedError:  service.ErrLoginAlreadyExists.Error(),
		},
		{
			name:           "неверный метод (GET)",
			method:         http.MethodGet,
			requestBody:    nil,
			setupMock:      func(m *mock.MockAuthService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedHeader: "",
			expectedError:  http.StatusText(http.StatusMethodNotAllowed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mock.MockAuthService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewAuthHandler(mockService)

			var req *http.Request
			if tt.requestBody != nil {
				req = httptest.NewRequest(tt.method, "/api/user/register", bytes.NewReader(tt.requestBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/api/user/register", nil)
			}

			w := httptest.NewRecorder()
			handler.RegisterHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Authorization"))
			}

			if tt.expectedStatus != http.StatusOK {
				var errResp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, errResp["error"])
				}
			}
		})
	}
}

func TestAuthHandler_LoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    []byte
		setupMock      func(*mock.MockAuthService)
		expectedStatus int
		expectedHeader string
		expectedError  string
	}{
		{
			name:   "успешная аутентификация",
			method: http.MethodPost,
			requestBody: func() []byte {
				b, _ := json.Marshal(model.RequestAuth{
					Login:    "testuser",
					Password: "password123",
				})
				return b
			}(),
			setupMock: func(m *mock.MockAuthService) {
				m.Token = "valid.jwt.token"
				m.ShouldFail = false
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "Bearer valid.jwt.token",
		},
		{
			name:           "неверный формат JSON",
			method:         http.MethodPost,
			requestBody:    []byte(`{"login": "test",`),
			setupMock:      func(m *mock.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
			expectedError:  http.StatusText(http.StatusBadRequest),
		},
		{
			name:   "неверный логин или пароль",
			method: http.MethodPost,
			requestBody: func() []byte {
				b, _ := json.Marshal(model.RequestAuth{
					Login:    "existing",
					Password: "pass123",
				})
				return b
			}(),
			setupMock: func(m *mock.MockAuthService) {
				m.ShouldFail = true
				m.FailWith = service.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
			expectedHeader: "",
			expectedError:  "invalid login or password",
		},
		{
			name:           "неверный метод (GET)",
			method:         http.MethodGet,
			requestBody:    nil,
			setupMock:      func(m *mock.MockAuthService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedHeader: "",
			expectedError:  http.StatusText(http.StatusMethodNotAllowed),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mock.MockAuthService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewAuthHandler(mockService)

			var req *http.Request
			if tt.requestBody != nil {
				req = httptest.NewRequest(tt.method, "/api/user/login", bytes.NewReader(tt.requestBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/api/user/login", nil)
			}

			w := httptest.NewRecorder()
			handler.LoginHandler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Authorization"))
			}

			if tt.expectedStatus != http.StatusOK {
				var errResp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
				if tt.expectedError != "" {
					assert.Equal(t, tt.expectedError, errResp["error"])
				}
			}
		})
	}
}
