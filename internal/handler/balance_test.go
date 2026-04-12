package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/domain"
	"gophermart/internal/handler"
	mw "gophermart/internal/middleware"
)

// mockBalanceProvider is a test double for handler.BalanceProvider.
type mockBalanceProvider struct {
	getBalanceFn func(ctx context.Context, userID int64) (*domain.Balance, error)
	withdrawFn   func(ctx context.Context, userID int64, orderNumber string, sum float64) error
}

func (m *mockBalanceProvider) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	return m.getBalanceFn(ctx, userID)
}
func (m *mockBalanceProvider) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	return m.withdrawFn(ctx, userID, orderNumber, sum)
}

// TestGetBalanceHandler_OK verifies 200 with JSON balance.
func TestGetBalanceHandler_OK(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{
		getBalanceFn: func(_ context.Context, _ int64) (*domain.Balance, error) {
			return &domain.Balance{Current: 500.5, Withdrawn: 42}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.GetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var b domain.Balance
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &b))
	assert.Equal(t, 500.5, b.Current)
	assert.Equal(t, 42.0, b.Withdrawn)
}

// TestGetBalanceHandler_InternalError verifies 500 on repository error.
func TestGetBalanceHandler_InternalError(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{
		getBalanceFn: func(_ context.Context, _ int64) (*domain.Balance, error) {
			return nil, assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.GetBalance(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func withdrawBody(order string, sum float64) *bytes.Buffer {
	b, _ := json.Marshal(map[string]interface{}{"order": order, "sum": sum})
	return bytes.NewBuffer(b)
}

// TestWithdrawHandler_OK verifies 200 on successful withdrawal.
func TestWithdrawHandler_OK(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{
		withdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error { return nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("2377225624", 100))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestWithdrawHandler_InsufficientBalance verifies 402 on insufficient funds.
func TestWithdrawHandler_InsufficientBalance(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{
		withdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return domain.ErrInsufficientBalance
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("2377225624", 999999))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusPaymentRequired, rr.Code)
}

// TestWithdrawHandler_InvalidOrder verifies 422 on invalid order number.
func TestWithdrawHandler_InvalidOrder(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{
		withdrawFn: func(_ context.Context, _ int64, _ string, _ float64) error {
			return domain.ErrInvalidOrderNumber
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", withdrawBody("1234567890", 10))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

// TestWithdrawHandler_BadRequest verifies 400 on malformed JSON.
func TestWithdrawHandler_BadRequest(t *testing.T) {
	h := handler.NewBalanceHandler(&mockBalanceProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString("bad"))
	req = req.WithContext(mw.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.Withdraw(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
