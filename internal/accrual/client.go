// Package accrual provides an HTTP client and background poller for the
// external loyalty points accrual calculation service.
package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// OrderResult holds the response from the accrual service for a given order.
type OrderResult struct {
	// Order is the order number.
	Order string `json:"order"`
	// Status is one of: REGISTERED, INVALID, PROCESSING, PROCESSED.
	Status string `json:"status"`
	// Accrual is the number of points awarded (absent unless status is PROCESSED).
	Accrual *float64 `json:"accrual,omitempty"`
}

// ErrRateLimit is returned when the accrual service responds with 429.
type ErrRateLimit struct {
	// RetryAfter is the duration to wait before the next request.
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e *ErrRateLimit) Error() string {
	return fmt.Sprintf("accrual rate limit: retry after %s", e.RetryAfter)
}

// ErrNotRegistered is returned when the accrual service responds with 204
// (the order is not yet known to the accrual system).
var ErrNotRegistered = fmt.Errorf("order not registered in accrual system")

// Client is an HTTP client for the accrual calculation service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new accrual Client targeting the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetOrder fetches the accrual status for the given order number.
// Returns ErrRateLimit on 429 and ErrNotRegistered on 204.
func (c *Client) GetOrder(ctx context.Context, number string) (*OrderResult, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result OrderResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &result, nil

	case http.StatusNoContent:
		return nil, ErrNotRegistered

	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, &ErrRateLimit{RetryAfter: retryAfter}

	default:
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

// parseRetryAfter parses the Retry-After header value (seconds integer).
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 60 * time.Second
	}
	secs, err := strconv.Atoi(header)
	if err != nil || secs <= 0 {
		return 60 * time.Second
	}
	return time.Duration(secs) * time.Second
}
