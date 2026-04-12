package service

import (
	"context"

	"gophermart/internal/domain"
	"gophermart/internal/luhn"
	"gophermart/internal/repository"
)

// BalanceService handles loyalty point balance queries and withdrawals.
type BalanceService struct {
	balanceRepo repository.BalanceRepo
}

// NewBalanceService creates a new BalanceService with the given repository.
func NewBalanceService(balanceRepo repository.BalanceRepo) *BalanceService {
	return &BalanceService{balanceRepo: balanceRepo}
}

// GetBalance returns the current balance for the user.
func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	return s.balanceRepo.GetBalance(ctx, userID)
}

// Withdraw deducts sum points from the user's balance for the given order number.
// Returns domain.ErrInvalidOrderNumber if orderNumber fails the Luhn check.
// Returns domain.ErrInsufficientBalance if current balance < sum.
func (s *BalanceService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	if !luhn.Valid(orderNumber) {
		return domain.ErrInvalidOrderNumber
	}
	return s.balanceRepo.Withdraw(ctx, userID, orderNumber, sum)
}
