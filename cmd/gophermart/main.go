// Package main is the entry point for the Gophermart loyalty system service.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"gophermart/internal/accrual"
	"gophermart/internal/admin"
	"gophermart/internal/config"
	"gophermart/internal/handler"
	"gophermart/internal/repository/postgres"
	"gophermart/internal/service"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic("init logger: " + err.Error())
	}
	defer log.Sync()

	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Connect to PostgreSQL and run migrations.
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURI)
	if err != nil {
		log.Fatal("connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := postgres.RunMigrations(pool); err != nil {
		log.Fatal("run migrations", zap.Error(err))
	}

	// Build repositories.
	userRepo := postgres.NewUserRepository(pool)
	orderRepo := postgres.NewOrderRepository(pool)
	balanceRepo := postgres.NewBalanceRepository(pool)
	withdrawalRepo := postgres.NewWithdrawalRepository(pool)

	// Build services.
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	orderSvc := service.NewOrderService(orderRepo)
	balanceSvc := service.NewBalanceService(balanceRepo)
	withdrawalSvc := service.NewWithdrawalService(withdrawalRepo)

	// Build HTTP handlers.
	authHandler := handler.NewAuthHandler(authSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	balanceHandler := handler.NewBalanceHandler(balanceSvc)
	withdrawalHandler := handler.NewWithdrawalHandler(withdrawalSvc)

	userRouter := handler.NewRouter(authHandler, orderHandler, balanceHandler, withdrawalHandler, authSvc, log)

	// Build admin panel.
	adminRepo := admin.NewPostgresRepo(pool)
	adminSvc := admin.NewService(cfg.AdminLogin, cfg.AdminPassword, cfg.JWTSecret)
	adminHandler := admin.NewHandler(adminSvc, adminRepo, log)
	adminRouter := admin.NewRouter(adminHandler, adminSvc)

	// Mount user API and admin panel on a single top-level mux.
	mux := http.NewServeMux()
	mux.Handle("/admin/", http.StripPrefix("/admin", adminRouter))
	mux.Handle("/", userRouter)

	// Start accrual poller in background.
	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClientWithLogger(cfg.AccrualSystemAddress, log)
		poller := accrual.NewPoller(accrualClient, orderRepo, log)
		go poller.Run(ctx)
	} else {
		log.Warn("ACCRUAL_SYSTEM_ADDRESS not set; accrual poller disabled")
	}

	// Start HTTP server with graceful shutdown.
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: mux,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting server", zap.String("addr", cfg.RunAddress))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Error("server error", zap.Error(err))
	case <-ctx.Done():
	}
	log.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown", zap.Error(err))
	}
}
