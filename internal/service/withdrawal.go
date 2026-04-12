package service

import (
	"context"

	"gophermart/internal/domain"
	"gophermart/internal/repository"
)

// WithdrawalService handles retrieval of withdrawal history.
type WithdrawalService struct {
	withdrawalRepo repository.WithdrawalRepo
}

// NewWithdrawalService creates a new WithdrawalService with the given repository.
func NewWithdrawalService(withdrawalRepo repository.WithdrawalRepo) *WithdrawalService {
	return &WithdrawalService{withdrawalRepo: withdrawalRepo}
}

// ListWithdrawals returns all withdrawals for the user, sorted newest-first.
func (s *WithdrawalService) ListWithdrawals(ctx context.Context, userID int64) ([]*domain.Withdrawal, error) {
	return s.withdrawalRepo.GetWithdrawalsByUserID(ctx, userID)
}
