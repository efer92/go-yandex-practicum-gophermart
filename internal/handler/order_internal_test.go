package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/domain"
	"gophermart/internal/handler"
	mw "gophermart/internal/middleware"
)

// TestUploadOrder_InternalError verifies 500 on unexpected repository error.
func TestUploadOrder_InternalError(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		submitFn: func(_ context.Context, _ int64, _ string) (*domain.Order, error) {
			return nil, assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.UploadOrder(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestListOrders_InternalError verifies 500 on repository error.
func TestListOrders_InternalError(t *testing.T) {
	h := handler.NewOrderHandler(&mockOrderProvider{
		listFn: func(_ context.Context, _ int64) ([]*domain.Order, error) {
			return nil, assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.ListOrders(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
