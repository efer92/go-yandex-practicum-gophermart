package accrual_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/accrual"
)

// TestGetOrder_OK verifies successful parsing of a 200 response.
func TestGetOrder_OK(t *testing.T) {
	accrualVal := 500.0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/12345678903", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accrual.OrderResult{
			Order:   "12345678903",
			Status:  "PROCESSED",
			Accrual: &accrualVal,
		})
	}))
	defer srv.Close()

	client := accrual.NewClient(srv.URL)
	result, err := client.GetOrder(context.Background(), "12345678903")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "PROCESSED", result.Status)
	assert.Equal(t, 500.0, *result.Accrual)
}

// TestGetOrder_NotRegistered verifies ErrNotRegistered on 204.
func TestGetOrder_NotRegistered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := accrual.NewClient(srv.URL)
	_, err := client.GetOrder(context.Background(), "12345678903")
	assert.ErrorIs(t, err, accrual.ErrNotRegistered)
}

// TestGetOrder_RateLimit verifies ErrRateLimit on 429 with Retry-After header.
func TestGetOrder_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := accrual.NewClient(srv.URL)
	_, err := client.GetOrder(context.Background(), "12345678903")

	require.Error(t, err)
	var rateLimitErr *accrual.ErrRateLimit
	require.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 30, int(rateLimitErr.RetryAfter.Seconds()))
}

// TestGetOrder_ServerError verifies an error is returned on 500.
func TestGetOrder_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := accrual.NewClient(srv.URL)
	_, err := client.GetOrder(context.Background(), "12345678903")
	assert.Error(t, err)
}

// TestGetOrder_RateLimit_NoHeader verifies default retry duration when Retry-After is absent.
func TestGetOrder_RateLimit_NoHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := accrual.NewClient(srv.URL)
	_, err := client.GetOrder(context.Background(), "12345678903")

	var rateLimitErr *accrual.ErrRateLimit
	require.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 60, int(rateLimitErr.RetryAfter.Seconds()))
}
