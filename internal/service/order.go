package service

import (
	"context"

	"gophermart/internal/domain"
	"gophermart/internal/luhn"
	"gophermart/internal/repository"
)

// OrderService handles order submission and retrieval.
type OrderService struct {
	orderRepo repository.OrderRepo
}

// NewOrderService creates a new OrderService with the given repository.
func NewOrderService(orderRepo repository.OrderRepo) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

// SubmitOrder validates and records a new order for the user.
// Returns domain.ErrInvalidOrderNumber if the number fails the Luhn check.
// Returns domain.ErrOrderAlreadyOwnedByUser if already submitted by this user.
// Returns domain.ErrOrderAlreadyOwnedByOther if submitted by a different user.
func (s *OrderService) SubmitOrder(ctx context.Context, userID int64, number string) (*domain.Order, error) {
	if !luhn.Valid(number) {
		return nil, domain.ErrInvalidOrderNumber
	}
	return s.orderRepo.CreateOrder(ctx, userID, number)
}

// ListOrders returns all orders for the user, sorted newest-first.
func (s *OrderService) ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return s.orderRepo.GetOrdersByUserID(ctx, userID)
}
