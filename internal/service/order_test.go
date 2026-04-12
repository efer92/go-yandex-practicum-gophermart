package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/domain"
	"gophermart/internal/service"
)

// mockOrderRepo is a test double for repository.OrderRepo.
type mockOrderRepo struct {
	createFn         func(ctx context.Context, userID int64, number string) (*domain.Order, error)
	getByUserFn      func(ctx context.Context, userID int64) ([]*domain.Order, error)
	getPendingFn     func(ctx context.Context, limit int) ([]*domain.Order, error)
	updateStatusFn   func(ctx context.Context, number string, status domain.OrderStatus, accrual *float64) error
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, userID int64, number string) (*domain.Order, error) {
	return m.createFn(ctx, userID, number)
}
func (m *mockOrderRepo) GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return m.getByUserFn(ctx, userID)
}
func (m *mockOrderRepo) GetPendingOrders(ctx context.Context, limit int) ([]*domain.Order, error) {
	return m.getPendingFn(ctx, limit)
}
func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, number string, status domain.OrderStatus, accrual *float64) error {
	return m.updateStatusFn(ctx, number, status, accrual)
}

// TestSubmitOrder_InvalidLuhn verifies ErrInvalidOrderNumber on a bad number.
func TestSubmitOrder_InvalidLuhn(t *testing.T) {
	svc := service.NewOrderService(&mockOrderRepo{})
	_, err := svc.SubmitOrder(context.Background(), 1, "1234567890") // invalid luhn
	assert.ErrorIs(t, err, domain.ErrInvalidOrderNumber)
}

// TestSubmitOrder_Success verifies a valid order is passed to the repository.
func TestSubmitOrder_Success(t *testing.T) {
	called := false
	repo := &mockOrderRepo{
		createFn: func(_ context.Context, userID int64, number string) (*domain.Order, error) {
			called = true
			return &domain.Order{ID: 1, UserID: userID, Number: number, Status: domain.OrderStatusNew}, nil
		},
	}
	svc := service.NewOrderService(repo)

	order, err := svc.SubmitOrder(context.Background(), 1, "12345678903")
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "12345678903", order.Number)
}

// TestSubmitOrder_AlreadyOwnedByUser verifies the correct sentinel error.
func TestSubmitOrder_AlreadyOwnedByUser(t *testing.T) {
	repo := &mockOrderRepo{
		createFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return nil, domain.ErrOrderAlreadyOwnedByUser
		},
	}
	svc := service.NewOrderService(repo)
	_, err := svc.SubmitOrder(context.Background(), 1, "12345678903")
	assert.ErrorIs(t, err, domain.ErrOrderAlreadyOwnedByUser)
}

// TestSubmitOrder_AlreadyOwnedByOther verifies the correct sentinel error.
func TestSubmitOrder_AlreadyOwnedByOther(t *testing.T) {
	repo := &mockOrderRepo{
		createFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return nil, domain.ErrOrderAlreadyOwnedByOther
		},
	}
	svc := service.NewOrderService(repo)
	_, err := svc.SubmitOrder(context.Background(), 1, "12345678903")
	assert.ErrorIs(t, err, domain.ErrOrderAlreadyOwnedByOther)
}

// TestListOrders verifies orders are returned from the repository.
func TestListOrders(t *testing.T) {
	expected := []*domain.Order{{ID: 1, Number: "12345678903"}}
	repo := &mockOrderRepo{
		getByUserFn: func(_ context.Context, _ int64) ([]*domain.Order, error) {
			return expected, nil
		},
	}
	svc := service.NewOrderService(repo)
	orders, err := svc.ListOrders(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, expected, orders)
}
