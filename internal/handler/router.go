package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"gophermart/internal/middleware"
)

// NewRouter assembles and returns the chi router with all routes and middleware.
func NewRouter(
	authHandler *AuthHandler,
	orderHandler *OrderHandler,
	balanceHandler *BalanceHandler,
	withdrawalHandler *WithdrawalHandler,
	tokenValidator middleware.TokenValidator,
	log *zap.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Compress)
	r.Use(middleware.Logger(log))

	// Public routes
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(tokenValidator))

		r.Post("/api/user/orders", orderHandler.UploadOrder)
		r.Get("/api/user/orders", orderHandler.ListOrders)

		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.Withdraw)

		r.Get("/api/user/withdrawals", withdrawalHandler.ListWithdrawals)
	})

	return r
}
