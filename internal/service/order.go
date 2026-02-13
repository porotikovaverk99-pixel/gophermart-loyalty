// Package service реализует бизнес-логику системы лояльности.
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/client"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/model"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/validator"
	"go.uber.org/zap"
)

var (
	ErrNumberAlreadyExists = errors.New("number already exists")

	ErrInvalidOrderNumber    = errors.New("invalid order number")
	ErrOrderBelongsToAnother = errors.New("number belongs to another user")
)

const (
	defaultTaskTimeout = 30 * time.Second
	schedulerInterval  = 10 * time.Second
)

// OrderService управляет заказами и их проверкой в accrual-системе.
type OrderService struct {
	repo           repository.OrderRepository
	accrual        *client.AccrualClient
	balanceService *BalanceService
	logger         *zap.Logger
	statusQueue    chan model.Order
	statusWorkers  int
	wg             sync.WaitGroup
	taskTimeout    time.Duration
	accrualQueue   chan model.AccrualTask
	accrualWorkers int
	stopChan       chan struct{}
}

// NewOrderService создает новый сервис заказов.
func NewOrderService(
	repo repository.OrderRepository,
	accrual *client.AccrualClient,
	balanceService *BalanceService,
	logger *zap.Logger,
	queueSize, statusWorkers, accrualWorkers int,
) *OrderService {
	return &OrderService{
		repo:           repo,
		accrual:        accrual,
		balanceService: balanceService,
		logger:         logger,
		statusQueue:    make(chan model.Order, queueSize),
		statusWorkers:  statusWorkers,
		taskTimeout:    defaultTaskTimeout,
		accrualQueue:   make(chan model.AccrualTask, queueSize),
		accrualWorkers: accrualWorkers,
	}
}

// UploadOrder загружает новый заказ пользователя.
// Возвращает ErrNumberAlreadyExists, если номер заказа уже загружен.
// Возвращает ErrOrderBelongsToAnother, если номер заказа уже загружен другим пользователем.
func (s *OrderService) UploadOrder(ctx context.Context, userID int64, number string) (int64, error) {

	if !validator.Luhn(number) {
		return 0, ErrInvalidOrderNumber
	}

	orderID, err := s.repo.CreateOrder(ctx, userID, number)
	if err != nil {
		if errors.Is(err, repository.ErrNumberAlreadyExists) {
			order, getErr := s.repo.GetOrderByNumber(ctx, number)
			if getErr != nil {
				return 0, fmt.Errorf("get existing order: %w", getErr)
			}

			if order.UserID == userID {
				return 0, ErrNumberAlreadyExists
			}
			return 0, ErrOrderBelongsToAnother
		}
		return 0, fmt.Errorf("create order: %w", err)
	}

	return orderID, nil
}

// GetUserOrders возвращает все заказы пользователя.
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {

	orders, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	return orders, nil
}

// StartAllWorkers запускает все воркеры для обработки заказов и начисления баллов.
func (s *OrderService) StartAllWorkers() {
	stopChan := make(chan struct{})
	s.stopChan = stopChan

	s.startStatusWorkers()
	s.startAccrualWorkers()

	go s.scheduler(stopChan)
}

// Stop останавливает все воркеры и ожидает их завершения.
func (s *OrderService) Stop() {
	close(s.stopChan)
	close(s.statusQueue)
	close(s.accrualQueue)
	s.wg.Wait()
}

func (s *OrderService) startStatusWorkers() {

	for i := 0; i < s.statusWorkers; i++ {
		s.wg.Add(1)
		go s.statusWorker(i)
	}

}

func (s *OrderService) startAccrualWorkers() {

	for i := 0; i < s.accrualWorkers; i++ {
		s.wg.Add(1)
		go s.accrualWorker(i)
	}

}

func (s *OrderService) statusWorker(workerID int) {

	defer s.wg.Done()

	for {
		select {

		case task, ok := <-s.statusQueue:
			if !ok {
				return
			}

			taskCtx, cancel := context.WithTimeout(context.Background(), s.taskTimeout)
			defer cancel()

			s.processOrder(taskCtx, task)

		}
	}
}

func (s *OrderService) accrualWorker(workerID int) {

	defer s.wg.Done()

	for {
		select {

		case task, ok := <-s.accrualQueue:
			if !ok {
				return
			}

			taskCtx, cancel := context.WithTimeout(context.Background(), s.taskTimeout)
			defer cancel()

			_ = s.balanceService.CreateAccrual(taskCtx, task.UserID, task.OrderNum, task.Amount)

		}
	}
}

func (s *OrderService) scheduler(stopChan chan struct{}) {

	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			taskCtx, cancel := context.WithTimeout(context.Background(), s.taskTimeout)
			defer cancel()
			orders, _ := s.repo.GetOrdersToProcess(taskCtx)
			for _, order := range orders {
				select {
				case s.statusQueue <- order:
				default:

				}
			}
		}
	}
}

func (s *OrderService) processOrder(ctx context.Context, order model.Order) {

	now := time.Now()
	s.repo.UpdateLastChecked(ctx, order.ID, now)

	resp, clientErr := s.accrual.GetOrder(order.Number)

	if clientErr != nil {

		if errors.Is(clientErr, client.ErrOrderNotRegistered) {
			s.logger.Info("Order not found in accrual, registering...",
				zap.String("order", order.Number))

			if regErr := s.accrual.RegisterOrder(order.Number); regErr != nil {
				s.logger.Error("Failed to register order in accrual",
					zap.String("order", order.Number),
					zap.Error(regErr))

				clientErr = regErr
			} else {
				s.logger.Info("Order registered in accrual successfully",
					zap.String("order", order.Number))
			}
		}

		nextRetryCount := order.RetryCount + 1
		nextCheck := s.calculateNextCheck(nextRetryCount, clientErr)
		s.repo.ScheduleNextCheck(ctx, order.ID, nextCheck, nextRetryCount)
		return
	}

	err := s.repo.UpdateOrderStatus(ctx, order.ID, resp.Status, resp.Accrual)
	if err != nil {
		s.logger.Error("Failed to update order status",
			zap.String("number", order.Number),
			zap.Error(err))
	} else {
		s.logger.Info("Status updated",
			zap.String("number", order.Number),
			zap.String("status", resp.Status),
			zap.Any("accrual", resp.Accrual))
	}

	if resp.Status == "PROCESSED" || resp.Status == "INVALID" {

		if resp.Status == "PROCESSED" && resp.Accrual != nil && order.Status != "PROCESSED" {
			s.notifyAccrual(ctx, order.UserID, order.Number, *resp.Accrual)
		}

		if err := s.repo.MarkOrderAsFinal(ctx, order.ID); err != nil {
			s.logger.Error("Failed to mark order as final",
				zap.String("number", order.Number),
				zap.Error(err))
		}
		return
	}

	nextCheck := s.calculateNextCheck(0, nil)
	s.repo.ScheduleNextCheck(ctx, order.ID, nextCheck, 0)
}

func (s *OrderService) notifyAccrual(ctx context.Context, userID int64, orderNum string, amount float64) {
	select {
	case s.accrualQueue <- model.AccrualTask{UserID: userID, OrderNum: orderNum, Amount: amount}:
	default:
		if err := s.balanceService.CreateAccrual(ctx, userID, orderNum, amount); err != nil {
			s.logger.Error("Failed to create accrual",
				zap.String("order", orderNum),
				zap.Error(err))
		}
	}
}

func (s *OrderService) calculateNextCheck(retryCount int, lastErr error) time.Time {

	if lastErr != nil {

		if errors.Is(lastErr, client.ErrRateLimitExceeded) {
			return time.Now().Add(60 * time.Second)
		}

		if errors.Is(lastErr, client.ErrAccrualUnavailable) {

			delays := []time.Duration{
				10 * time.Second,
				30 * time.Second,
				1 * time.Minute,
				2 * time.Minute,
				5 * time.Minute,
			}

			if retryCount >= len(delays) {
				retryCount = len(delays) - 1
			}

			return time.Now().Add(delays[retryCount])
		}
	}

	delays := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
		1 * time.Minute,
		2 * time.Minute,
		5 * time.Minute,
	}

	if retryCount >= len(delays) {
		retryCount = len(delays) - 1
	}

	return time.Now().Add(delays[retryCount])
}
