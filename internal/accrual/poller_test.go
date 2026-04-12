package accrual_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"gophermart/internal/accrual"
	"gophermart/internal/domain"
)

// mockOrderRepoForPoller is a test double for repository.OrderRepo used in poller tests.
type mockOrderRepoForPoller struct {
	pendingOrders []*domain.Order
	updated       []string
}

func (m *mockOrderRepoForPoller) CreateOrder(_ context.Context, _ int64, _ string) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepoForPoller) GetOrdersByUserID(_ context.Context, _ int64) ([]*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepoForPoller) GetPendingOrders(_ context.Context, _ int) ([]*domain.Order, error) {
	return m.pendingOrders, nil
}
func (m *mockOrderRepoForPoller) UpdateOrderStatus(_ context.Context, number string, _ domain.OrderStatus, _ *float64) error {
	m.updated = append(m.updated, number)
	return nil
}

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

	repo := &mockOrderRepoForPoller{
		pendingOrders: []*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		},
	}

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		poller.Run(ctx)
		close(done)
	}()

	// Wait until the order is updated or timeout.
	deadline := time.After(2 * time.Second)
	for {
		if len(repo.updated) > 0 {
			break
		}
		select {
		case <-deadline:
			t.Error("timed out waiting for order update")
			cancel()
			<-done
			return
		default:
			time.Sleep(50 * time.Millisecond)
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

	repo := &mockOrderRepoForPoller{
		pendingOrders: []*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		},
	}

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

	repo := &mockOrderRepoForPoller{
		pendingOrders: []*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		},
	}

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	poller.Run(ctx)

	if len(repo.updated) > 0 {
		t.Errorf("expected no updates but got %d", len(repo.updated))
	}
}

// TestPoller_HandlesServerError verifies the poller skips orders on 500.
func TestPoller_HandlesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	repo := &mockOrderRepoForPoller{
		pendingOrders: []*domain.Order{
			{ID: 1, Number: "12345678903", Status: domain.OrderStatusNew},
		},
	}

	client := accrual.NewClient(srv.URL)
	poller := accrual.NewPoller(client, repo, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	poller.Run(ctx)

	if len(repo.updated) > 0 {
		t.Errorf("expected no updates but got %d", len(repo.updated))
	}
}
