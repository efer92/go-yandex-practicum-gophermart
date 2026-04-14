package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/domain"
	"gophermart/internal/service"
)

// mockBalanceRepo is a test double for repository.BalanceRepo.
type mockBalanceRepo struct {
	getBalanceFn func(ctx context.Context, userID int64) (*domain.Balance, error)
	withdrawFn   func(ctx context.Context, userID int64, orderNumber string, sum float64) error
}

func (m *mockBalanceRepo) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	return m.getBalanceFn(ctx, userID)
}
func (m *mockBalanceRepo) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	return m.withdrawFn(ctx, userID, orderNumber, sum)
}

// TestGetBalance verifies the balance is returned from the repository.
func TestGetBalance(t *testing.T) {
	expected := &domain.Balance{Current: 100, Withdrawn: 50}
	repo := &mockBalanceRepo{
		getBalanceFn: func(_ context.Context, _ int64) (*domain.Balance, error) {
			return expected, nil
		},
	}
	svc := service.NewBalanceService(repo)
	bal, err := svc.GetBalance(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, expected, bal)
}

// TestWithdraw_Success verifies a valid withdrawal is passed to the repository.
func TestWithdraw_Success(t *testing.T) {
	called := false
	repo := &mockBalanceRepo{
		withdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			called = true
			return nil
		},
	}
	svc := service.NewBalanceService(repo)
	err := svc.Withdraw(context.Background(), 1, "2377225624", 50)
	require.NoError(t, err)
	assert.True(t, called)
}

// TestWithdraw_InvalidLuhn verifies ErrInvalidOrderNumber on bad order number.
func TestWithdraw_InvalidLuhn(t *testing.T) {
	repo := &mockBalanceRepo{}
	svc := service.NewBalanceService(repo)
	err := svc.Withdraw(context.Background(), 1, "1234567890", 50)
	assert.ErrorIs(t, err, domain.ErrInvalidOrderNumber)
}

// TestWithdraw_InsufficientBalance verifies ErrInsufficientBalance is propagated.
func TestWithdraw_InsufficientBalance(t *testing.T) {
	repo := &mockBalanceRepo{
		withdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return domain.ErrInsufficientBalance
		},
	}
	svc := service.NewBalanceService(repo)
	err := svc.Withdraw(context.Background(), 1, "2377225624", 999999)
	assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
}
