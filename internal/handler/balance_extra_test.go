package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/domain"
	"gophermart/internal/handler"
	mw "gophermart/internal/middleware"
)

// TestWithdrawHandler_InternalError verifies 500 on unexpected withdrawal error.
func TestWithdrawHandler_InternalError(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{
		withdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("2377225624", 10))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestWithdrawHandler_ZeroSum verifies 400 when sum is zero.
func TestWithdrawHandler_ZeroSum(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("2377225624", 0))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestWithdrawHandler_NegativeSum verifies 400 when sum is negative.
func TestWithdrawHandler_NegativeSum(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("2377225624", -1))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestWithdrawHandler_EmptyOrder verifies 400 when order is empty string.
func TestWithdrawHandler_EmptyOrder(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("", 10))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// mockBalanceProviderNoWithdraw implements BalanceProvider with no withdraw fn.
type mockBalanceProviderNoWithdraw struct{}

func (m *mockBalanceProviderNoWithdraw) GetBalance(_ context.Context, _ int64) (*domain.Balance, error) {
	return nil, assert.AnError
}
func (m *mockBalanceProviderNoWithdraw) Withdraw(_ context.Context, _ int64, _ string, _ float64) error {
	return nil
}
