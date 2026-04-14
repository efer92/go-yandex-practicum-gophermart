package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/domain"
	"gophermart/internal/service"
)

// mockWithdrawalRepo is a test double for repository.WithdrawalRepo.
type mockWithdrawalRepo struct {
	getByUserFn func(ctx context.Context, userID int64) ([]*domain.Withdrawal, error)
}

func (m *mockWithdrawalRepo) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*domain.Withdrawal, error) {
	return m.getByUserFn(ctx, userID)
}

// TestListWithdrawals verifies withdrawals are returned from the repository.
func TestListWithdrawals(t *testing.T) {
	expected := []*domain.Withdrawal{
		{OrderNumber: "2377225624", Sum: 100, ProcessedAt: time.Now()},
	}
	repo := &mockWithdrawalRepo{
		getByUserFn: func(_ context.Context, _ int64) ([]*domain.Withdrawal, error) {
			return expected, nil
		},
	}
	svc := service.NewWithdrawalService(repo)
	list, err := svc.ListWithdrawals(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, expected, list)
}

// TestListWithdrawals_Empty verifies empty slice is returned when no withdrawals exist.
func TestListWithdrawals_Empty(t *testing.T) {
	repo := &mockWithdrawalRepo{
		getByUserFn: func(_ context.Context, _ int64) ([]*domain.Withdrawal, error) {
			return nil, nil
		},
	}
	svc := service.NewWithdrawalService(repo)
	list, err := svc.ListWithdrawals(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, list)
}
