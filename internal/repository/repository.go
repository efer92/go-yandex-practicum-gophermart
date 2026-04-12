// Package repository defines the data access interfaces for the Gophermart system.
package repository

import (
	"context"

	"gophermart/internal/domain"
)

// UserRepo defines persistence operations for users.
type UserRepo interface {
	// CreateUser inserts a new user and returns the created record.
	// Returns domain.ErrUserAlreadyExists if the login is taken.
	CreateUser(ctx context.Context, login, passwordHash string) (*domain.User, error)

	// GetUserByLogin retrieves a user by their login name.
	// Returns nil, nil if not found.
	GetUserByLogin(ctx context.Context, login string) (*domain.User, error)
}

// OrderRepo defines persistence operations for loyalty orders.
type OrderRepo interface {
	// CreateOrder inserts a new order for the given user.
	// Returns domain.ErrOrderAlreadyOwnedByUser if the order belongs to userID.
	// Returns domain.ErrOrderAlreadyOwnedByOther if the order belongs to another user.
	CreateOrder(ctx context.Context, userID int64, number string) (*domain.Order, error)

	// GetOrdersByUserID returns all orders for a user, sorted newest-first.
	GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)

	// GetPendingOrders returns orders in NEW or PROCESSING state, up to limit.
	GetPendingOrders(ctx context.Context, limit int) ([]*domain.Order, error)

	// UpdateOrderStatus updates the status and optional accrual for an order.
	UpdateOrderStatus(ctx context.Context, number string, status domain.OrderStatus, accrual *float64) error
}

// BalanceRepo defines persistence operations for user balances.
type BalanceRepo interface {
	// GetBalance returns the current balance for the given user.
	// Current = total accrued (PROCESSED orders) - total withdrawn.
	GetBalance(ctx context.Context, userID int64) (*domain.Balance, error)

	// Withdraw atomically checks balance and records a withdrawal.
	// Returns domain.ErrInsufficientBalance if current < sum.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error
}

// WithdrawalRepo defines persistence operations for withdrawal records.
type WithdrawalRepo interface {
	// GetWithdrawalsByUserID returns all withdrawals for a user, sorted newest-first.
	GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*domain.Withdrawal, error)
}
