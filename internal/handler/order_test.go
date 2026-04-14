package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/domain"
	"gophermart/internal/handler"
	mw "gophermart/internal/middleware"
)

// mockOrderProvider is a test double for handler.OrderProvider.
type mockOrderProvider struct {
	submitFn func(ctx context.Context, userID int64, number string) (*domain.Order, error)
	listFn   func(ctx context.Context, userID int64) ([]*domain.Order, error)
}

func (m *mockOrderProvider) SubmitOrder(ctx context.Context, userID int64, number string) (*domain.Order, error) {
	return m.submitFn(ctx, userID, number)
}
func (m *mockOrderProvider) ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return m.listFn(ctx, userID)
}

// withUserID injects a user ID into the request context (simulates auth middleware).
func withUserID(r *http.Request, userID string) *http.Request {
	ctx := mw.WithUserID(r.Context(), userID)
	return r.WithContext(ctx)
}

// TestUploadOrder_Accepted verifies 202 on new valid order.
func TestUploadOrder_Accepted(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		submitFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return &domain.Order{}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.UploadOrder(rr, req)
	assert.Equal(t, http.StatusAccepted, rr.Code)
}

// TestUploadOrder_AlreadyOwned verifies 200 when same user re-submits an order.
func TestUploadOrder_AlreadyOwned(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		submitFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return nil, domain.ErrOrderAlreadyOwnedByUser
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.UploadOrder(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestUploadOrder_Conflict verifies 409 when a different user owns the order.
func TestUploadOrder_Conflict(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		submitFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return nil, domain.ErrOrderAlreadyOwnedByOther
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.UploadOrder(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

// TestUploadOrder_InvalidNumber verifies 422 on Luhn check failure.
func TestUploadOrder_InvalidNumber(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		submitFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return nil, domain.ErrInvalidOrderNumber
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("1234567890"))
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.UploadOrder(rr, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

// TestUploadOrder_EmptyBody verifies 400 on empty body.
func TestUploadOrder_EmptyBody(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(""))
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.UploadOrder(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestListOrders_OK verifies 200 and JSON body when orders exist.
func TestListOrders_OK(t *testing.T) {
	accrual := 500.0
	h := handler.NewOrderHandler(&mockOrderProvider{
		listFn: func(_ context.Context, _ int64) ([]*domain.Order, error) {
			return []*domain.Order{
				{Number: "12345678903", Status: domain.OrderStatusProcessed, Accrual: &accrual, UploadedAt: time.Now()},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.ListOrders(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
}

// TestListOrders_NoContent verifies 204 when no orders exist.
func TestListOrders_NoContent(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		listFn: func(_ context.Context, _ int64) ([]*domain.Order, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = withUserID(req, "1")
	rr := httptest.NewRecorder()
	h.ListOrders(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}
