package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/domain"
	"gophermart/internal/handler"
	mw "gophermart/internal/middleware"
)

// mockWithdrawalProvider is a test double for handler.WithdrawalProvider.
type mockWithdrawalProvider struct {
	listFn func(ctx context.Context, userID int64) ([]*domain.Withdrawal, error)
}

func (m *mockWithdrawalProvider) ListWithdrawals(ctx context.Context, userID int64) ([]*domain.Withdrawal, error) {
	return m.listFn(ctx, userID)
}

// TestListWithdrawalsHandler_OK verifies 200 with JSON body when withdrawals exist.
func TestListWithdrawalsHandler_OK(t *testing.T) {
	h := handler.NewWithdrawalHandler(&mockWithdrawalProvider{
		listFn: func(_ context.Context, _ int64) ([]*domain.Withdrawal, error) {
			return []*domain.Withdrawal{
				{OrderNumber: "2377225624", Sum: 500, ProcessedAt: time.Now()},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.ListWithdrawals(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
}

// TestListWithdrawalsHandler_NoContent verifies 204 when there are no withdrawals.
func TestListWithdrawalsHandler_NoContent(t *testing.T) {
	h := handler.NewWithdrawalHandler(&mockWithdrawalProvider{
		listFn: func(_ context.Context, _ int64) ([]*domain.Withdrawal, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.ListWithdrawals(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

// TestListWithdrawalsHandler_InternalError verifies 500 on repository error.
func TestListWithdrawalsHandler_InternalError(t *testing.T) {
	h := handler.NewWithdrawalHandler(&mockWithdrawalProvider{
		listFn: func(_ context.Context, _ int64) ([]*domain.Withdrawal, error) {
			return nil, assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(mw.WithUserID(req.Context(), "1"))
	rr := httptest.NewRecorder()
	h.ListWithdrawals(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
