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
	"go.uber.org/zap"
)

var (
	ErrNumberAlreadyExists = errors.New("number already exists")
)

type OrderService struct {
	repo        repository.OrderRepository
	accrual     *client.AccrualClient
	logger      *zap.Logger
	cancelFunc  context.CancelFunc
	jobQueue    chan model.Order
	workerCount int
	wg          sync.WaitGroup
	jobTimeout  time.Duration
}

func NewOrderService(
	repo repository.OrderRepository,
	accrual *client.AccrualClient,
	logger *zap.Logger,
	queueSize, workerCount int,
) *OrderService {
	return &OrderService{
		repo:        repo,
		accrual:     accrual,
		logger:      logger,
		jobQueue:    make(chan model.Order, queueSize),
		workerCount: workerCount,
		jobTimeout:  30 * time.Second,
	}
}

func (s *OrderService) UploadOrder(ctx context.Context, userID int64, number string) (int64, error) {

	orderID, err := s.repo.CreateOrder(ctx, userID, number)
	if err != nil {
		if errors.Is(err, repository.ErrNumberAlreadyExists) {
			return 0, ErrNumberAlreadyExists
		}
		return 0, fmt.Errorf("create order: %w", err)
	}

	return orderID, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int64) ([]model.Order, error) {

	orders, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}

	return orders, nil
}

func (s *OrderService) StartWorkers() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.taskWorker(ctx, i)
	}

	go s.scheduler(ctx)
}

func (s *OrderService) scheduler(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			orders, _ := s.repo.GetOrdersToProcess(ctx)
			for _, order := range orders {
				select {
				case s.jobQueue <- order:
				default:

				}
			}
		}
	}
}

func (s *OrderService) taskWorker(ctx context.Context, workerID int) {

	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case task, ok := <-s.jobQueue:
			if !ok {
				return
			}

			taskCtx, cancel := context.WithTimeout(context.Background(), s.jobTimeout)
			defer cancel()

			s.processOrder(taskCtx, task)

		}
	}
}

func (s *OrderService) processOrder(ctx context.Context, order model.Order) {

	now := time.Now()
	s.repo.UpdateLastChecked(ctx, order.ID, now)

	resp, err := s.accrual.GetOrder(order.Number)

	if err != nil {

		if errors.Is(err, client.ErrOrderNotRegistered) {

			s.logger.Info("Order not found in accrual, registering...",
				zap.String("order", order.Number))

			if regErr := s.accrual.RegisterOrder(order.Number); regErr != nil {
				s.logger.Error("Failed to register order in accrual",
					zap.String("order", order.Number),
					zap.Error(regErr))
			} else {
				s.logger.Info("Order registered in accrual successfully",
					zap.String("order", order.Number))
			}
		}

		nextCheck := s.calculateNextCheck(order.RetryCount + 1)
		s.repo.ScheduleNextCheck(ctx, order.ID, nextCheck, order.RetryCount+1)
		return
	}

	err = s.repo.UpdateOrderStatus(ctx, order.ID, resp.Status, resp.Accrual)
	if err != nil {
		s.logger.Error("Failed to update order status", zap.String("number", order.Number))
	} else {
		s.logger.Info("status updated", zap.String("number", order.Number))
	}

	if resp.Status == "PROCESSED" || resp.Status == "INVALID" {
		return
	}

	nextCheck := s.calculateNextCheck(0)
	s.repo.ScheduleNextCheck(ctx, order.ID, nextCheck, 0)

}

func (s *OrderService) calculateNextCheck(retryCount int) time.Time {
	baseDelay := 5 * time.Second
	delay := baseDelay * time.Duration(1<<uint(retryCount))
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	return time.Now().Add(delay)
}
