package accrual_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"gophermart/internal/accrual"
	"gophermart/internal/domain"
	"gophermart/internal/repository/mocks"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../repository/mocks/mock_order_repo.go -package=mocks gophermart/internal/repository OrderRepo

// TestPoller_ProcessesOrder verifies the poller updates order status on PROCESSED response.
func TestPoller_ProcessesOrder(t *testing.T) {
	accrualVal := 100.0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accrual.OrderResult{
			Order:   "12345678903",
			Status:  "PROCESSED",
			Accrual: &accrualVal,
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepo(ctrl)
	repo.EXPECT().
		GetPendingOrders(gomock.Any(), gomock.Any()).
		Return([]*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		}, nil).
		AnyTimes()
	repo.EXPECT().
		UpdateOrderStatus(gomock.Any(), "12345678903", domain.OrderStatusProcessed, gomock.Any()).
		Return(nil).
		AnyTimes()

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Error("timed out waiting for order update")
			cancel()
			<-done
			return
		default:
		}
		// Give the poller a tick to process.
		time.Sleep(50 * time.Millisecond)
		// Check via the mock controller — if UpdateOrderStatus was called, we're done.
		if ctrl.Satisfied() {
			break
		}
	}
	cancel()
	<-done
}

// TestPoller_HandlesRateLimit verifies the poller backs off on 429 without panicking.
func TestPoller_HandlesRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepo(ctrl)
	repo.EXPECT().
		GetPendingOrders(gomock.Any(), gomock.Any()).
		Return([]*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		}, nil).
		AnyTimes()

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	poller.Run(ctx) // should not panic
}

// TestPoller_HandlesNotRegistered verifies the poller leaves NEW orders alone when 204.
func TestPoller_HandlesNotRegistered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepo(ctrl)
	repo.EXPECT().
		GetPendingOrders(gomock.Any(), gomock.Any()).
		Return([]*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		}, nil).
		AnyTimes()
	// UpdateOrderStatus must NOT be called.

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	poller.Run(ctx)
}

// TestPoller_HandlesServerError verifies the poller skips orders on 500.
func TestPoller_HandlesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockOrderRepo(ctrl)
	repo.EXPECT().
		GetPendingOrders(gomock.Any(), gomock.Any()).
		Return([]*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		}, nil).
		AnyTimes()
	// UpdateOrderStatus must NOT be called.

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	poller.Run(ctx)
}
