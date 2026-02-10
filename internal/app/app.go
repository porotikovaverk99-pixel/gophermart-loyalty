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
	shutdownTimeout       = 15 * time.Second
	serverShutdownTimeout = 10 * time.Second
)

type App struct {
	config   *config.Config
	logger   *zap.Logger
	server   *server.Server
	clients  *Clients
	repos    *Repositories
	services *Services
	handlers *Handlers
}

type Clients struct {
	Accrual *client.AccrualClient
}

type Repositories struct {
	Users   repository.UserRepository
	Orders  repository.OrderRepository
	Balance repository.BalanceRepository
}

type Services struct {
	Auth    *service.AuthService
	Orders  *service.OrderService
	Balance *service.BalanceService
}

type Handlers struct {
	Auth    *handler.AuthHandler
	Orders  *handler.OrderHandler
	Balance *handler.BalanceHandler
}

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

	clients := &Clients{
		Accrual: client.NewAccrualClient("http://localhost:8081"),
	}

	repos := &Repositories{
		Users:   repository.NewUserRepository(db.GetPool()),
		Orders:  repository.NewOrderRepository(db.GetPool()),
		Balance: repository.NewBalanceRepository(db.GetPool()),
	}

	jwtManager := auth.NewJWTManager(cfg.SecretKey, 30*time.Minute)
	services := &Services{
		Auth:    service.NewAuthService(repos.Users, jwtManager),
		Orders:  service.NewOrderService(repos.Orders, clients.Accrual, zapLogger, cfg.WorkerQueueSize, cfg.WorkerCount),
		Balance: service.NewBalanceService(repos.Balance),
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

	JWTManager := auth.NewJWTManager("", 3*time.Hour)

	authMiddleware := auth.AuthMiddleware(JWTManager)

	a.server.Handle("/api/user/orders", authMiddleware(a.handlers.Orders.BaseOrderHandler()))

	a.server.Handle("/api/user/balance", authMiddleware(a.handlers.Balance.GetBalanceHandler()))
	a.server.Handle("/api/user/balance/withdraw", authMiddleware(a.handlers.Balance.BalanceWithdrawHandler()))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func (a *App) Run() error {

	a.services.Orders.StartWorkers()
	//defer a.services.Orders.StopWorkers()

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

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	a.logger.Info("Shutting down HTTP server...")
	serverCtx, serverCancel := context.WithTimeout(ctx, serverShutdownTimeout)
	defer serverCancel()

	if err := a.server.Shutdown(serverCtx); err != nil {
		a.logger.Error("HTTP server shutdown error", zap.Error(err))
	} else {
		a.logger.Info("HTTP server stopped gracefully")
	}

	//if err := a.repos.Close(); err != nil {
	//a.logger.Error("Failed to close database connections", zap.Error(err))
	//}

	select {
	case <-ctx.Done():
		a.logger.Warn("Shutdown timed out - some operations may not have completed")
	default:
		a.logger.Info("Graceful shutdown completed successfully")
	}
}

func (a *App) Close() {
	_ = a.logger.Sync()
}
