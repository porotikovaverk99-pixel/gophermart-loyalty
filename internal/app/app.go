// Package app инициализирует и запускает приложение Gophermart.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/auth"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/client"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/config"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/handler"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/logger"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/repository"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/server"
	"github.com/porotikovaverk99-pixel/gophermart-loyalty/internal/service"
	"go.uber.org/zap"
)

const (
	shutdownTimeout = 15 * time.Second
)

// App представляет основное приложение, объединяющее все компоненты.
type App struct {
	config   *config.Config
	logger   *zap.Logger
	server   *server.Server
	clients  *Clients
	repos    *Repositories
	services *Services
	handlers *Handlers
}

// Clients содержит HTTP-клиенты для внешних сервисов.
type Clients struct {
	Accrual *client.AccrualClient
}

// Repositories содержит все репозитории для работы с БД.
type Repositories struct {
	Users   repository.UserRepository
	Orders  repository.OrderRepository
	Balance repository.BalanceRepository
	db      *repository.Database
}

// Services содержит всю бизнес-логику приложения.
type Services struct {
	Auth    *service.AuthService
	Orders  *service.OrderService
	Balance *service.BalanceService
}

// Handlers содержит HTTP-обработчики.
type Handlers struct {
	Auth    *handler.AuthHandler
	Orders  *handler.OrderHandler
	Balance *handler.BalanceHandler
}

// NewApp создает и инициализирует новое приложение.
// Загружает конфигурацию, подключается к БД, инициализирует все зависимости.
func NewApp() (*App, error) {

	cfg := config.ParseFlags()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("initialize logger: %w", err)
	}
	zapLogger := logger.Log

	db, err := repository.NewDatabase(cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}

	clients := &Clients{
		Accrual: client.NewAccrualClient(cfg.AccrualSystemAddress),
	}

	repos := &Repositories{
		Users:   repository.NewUserRepository(db.GetPool()),
		Orders:  repository.NewOrderRepository(db.GetPool()),
		Balance: repository.NewBalanceRepository(db.GetPool()),
		db:      db,
	}

	BalanceService := service.NewBalanceService(repos.Balance)

	jwtManager := auth.NewJWTManager(cfg.SecretKey, cfg.JWTExpiry)
	services := &Services{
		Auth: service.NewAuthService(repos.Users, jwtManager),
		Orders: service.NewOrderService(
			repos.Orders,
			clients.Accrual,
			BalanceService,
			zapLogger,
			cfg.WorkerQueueSize,
			cfg.WorkerCount,
			cfg.WorkerCount,
		),
		Balance: BalanceService,
	}

	handlers := &Handlers{
		Auth:    handler.NewAuthHandler(services.Auth),
		Orders:  handler.NewOrderHandler(services.Orders),
		Balance: handler.NewBalanceHandler(services.Balance),
	}

	srv := server.New(cfg.RunAddr)

	app := &App{
		config:   &cfg,
		logger:   zapLogger,
		server:   srv,
		clients:  clients,
		repos:    repos,
		services: services,
		handlers: handlers,
	}

	app.setupRoutes()

	return app, nil
}

func (a *App) setupRoutes() {
	router := a.server.Router()

	a.server.Handle("/api/user/register", a.handlers.Auth.RegisterHandler())
	a.server.Handle("/api/user/login", a.handlers.Auth.LoginHandler())

	authMiddleware := auth.AuthMiddleware(a.services.Auth.GetManager())

	a.server.Handle("/api/user/orders", authMiddleware(a.handlers.Orders.BaseOrderHandler()))

	a.server.Handle("/api/user/balance", authMiddleware(a.handlers.Balance.GetBalanceHandler()))
	a.server.Handle("/api/user/balance/withdraw", authMiddleware(a.handlers.Balance.BalanceWithdrawHandler()))
	a.server.Handle("/api/user/withdrawals", authMiddleware(a.handlers.Balance.GetWithdrawalsHandler()))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// Run запускает HTTP-сервер и воркеры, ожидает сигналов завершения.
// Блокирует выполнение до остановки приложения.
func (a *App) Run() error {

	a.services.Orders.StartAllWorkers()

	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info("Starting HTTP server", zap.String("address", a.config.RunAddr))
		if err := a.server.Run(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("server error: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		a.logger.Error("Server stopped with error", zap.Error(err))
		a.shutdown()
		return err

	case sig := <-sigChan:
		a.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		a.shutdown()
		a.logger.Info("Application shutdown completed")
	}

	return nil
}

func (a *App) shutdown() {

	a.logger.Info("Starting graceful shutdown")

	a.logger.Info("Stopping order service workers...")
	a.services.Orders.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	a.logger.Info("Shutting down HTTP server...")
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("HTTP server shutdown error", zap.Error(err))
	} else {
		a.logger.Info("HTTP server stopped gracefully")
	}

	if a.repos.db != nil {
		a.logger.Info("Closing database connection...")
		a.repos.db.Close()
	}

	select {
	case <-ctx.Done():
		a.logger.Warn("Shutdown timed out - some operations may not have completed")
	default:
		a.logger.Info("Graceful shutdown completed successfully")
	}
}

// Close освобождает ресурсы приложения (логгер).
func (a *App) Close() {
	_ = a.logger.Sync()
}
