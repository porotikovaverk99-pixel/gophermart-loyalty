// Package client реализует HTTP-клиент для взаимодействия с внешним accrual-сервисом
// системы начисления баллов.
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var (
	// ErrOrderNotRegistered возвращается, когда заказ не найден в accrual-системе (204).
	ErrOrderNotRegistered = errors.New("order not registered in accrual system")

	// ErrRateLimitExceeded возвращается при превышении лимита запросов (429).
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrAccrualUnavailable возвращается при недоступности сервиса (5xx).
	ErrAccrualUnavailable = errors.New("accrual service unavailable")
)

// AccrualClient предоставляет методы для работы с внешним сервисом начисления баллов.
type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAccrualClient создает новый клиент для accrual-сервиса.
// baseURL: адрес сервиса (например, http://localhost:8080)
func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

// GetOrder запрашивает информацию о заказе в accrual-системе.
// GET /api/orders/{number}
// Возвращает статус заказа и сумму начисленных баллов.
// Ошибки: ErrOrderNotRegistered, ErrRateLimitExceeded, ErrAccrualUnavailable.
func (c *AccrualClient) GetOrder(orderNumber string) (*AccrualResponse, error) {

	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "gophermart-loyalty/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &result, nil

	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, fmt.Errorf("%w: retry after %v", ErrRateLimitExceeded, retryAfter)

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("%w: internal server error", ErrAccrualUnavailable)

	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

// AccrualResponse представляет ответ от accrual-сервиса.
type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"` // REGISTERED, PROCESSING, INVALID, PROCESSED
	Accrual *float64 `json:"accrual,omitempty"`
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 60 * time.Second
	}

	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	if t, err := time.Parse(time.RFC1123, header); err == nil {
		return time.Until(t)
	}

	return 60 * time.Second
}

// GetOrderWithRetry выполняет запрос с повторными попытками при ошибках 429.
// maxRetries: максимальное количество попыток.
func (c *AccrualClient) GetOrderWithRetry(orderNumber string, maxRetries int) (*AccrualResponse, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		resp, err := c.GetOrder(orderNumber)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if !errors.Is(err, ErrRateLimitExceeded) {
			return nil, err
		}

		sleepTime := time.Duration(1<<uint(i)) * time.Second
		if sleepTime > 30*time.Second {
			sleepTime = 30 * time.Second
		}
		time.Sleep(sleepTime)
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RegisterOrder регистрирует заказ в accrual-системе.
// POST /api/orders
// Body: {"order": "number"}
// Ошибки: ErrRateLimitExceeded, ошибки валидации номера заказа.
func (c *AccrualClient) RegisterOrder(orderNumber string) error {

	requestBody := map[string]string{
		"order": orderNumber,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/orders", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("order already registered by another user")
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("invalid order number format")
	case http.StatusTooManyRequests:
		return ErrRateLimitExceeded
	default:
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}
