package accrual

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"gophermart/internal/domain"
	"gophermart/internal/repository"
)

const (
	defaultPollInterval = time.Second
	defaultWorkers      = 3
	defaultBatchSize    = 50
)

// Poller periodically polls the accrual service for pending orders and
// updates their status in the repository.
type Poller struct {
	client     *Client
	orderRepo  repository.OrderRepo
	log        *zap.Logger
	pauseUntil atomic.Int64 // unix nano; workers stop sending when time.Now() is before this
}

// NewPoller creates a new Poller.
func NewPoller(client *Client, orderRepo repository.OrderRepo, log *zap.Logger) *Poller {
	return &Poller{
		client:    client,
		orderRepo: orderRepo,
		log:       log,
	}
}

// isPaused reports whether the poller is currently in a rate-limit backoff period.
func (p *Poller) isPaused() bool {
	until := p.pauseUntil.Load()
	return until > 0 && time.Now().UnixNano() < until
}

// setPause extends the rate-limit pause to at least now+d.
// If pauseUntil already holds a more distant moment, it is left unchanged.
func (p *Poller) setPause(d time.Duration) {
	newUntil := time.Now().Add(d).UnixNano()
	for {
		current := p.pauseUntil.Load()
		if current >= newUntil {
			return
		}
		if p.pauseUntil.CompareAndSwap(current, newUntil) {
			return
		}
	}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	sem := make(chan struct{}, defaultWorkers)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Skip the entire batch while rate-limited; individual workers
			// also check isPaused() so any already-dispatched ones stop early.
			if p.isPaused() {
				continue
			}

			orders, err := p.orderRepo.GetPendingOrders(ctx, defaultBatchSize)
			if err != nil {
				p.log.Error("get pending orders", zap.Error(err))
				continue
			}
			for _, order := range orders {
				order := order
				select {
				case sem <- struct{}{}:
					go func() {
						defer func() { <-sem }()
						p.processOrder(ctx, order)
					}()
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// processOrder fetches accrual info for one order and updates the DB.
// It is a no-op if the poller is currently rate-limited.
func (p *Poller) processOrder(ctx context.Context, order *domain.Order) {
	// A sibling worker may have set a pause after we were dispatched.
	if p.isPaused() {
		return
	}

	result, err := p.client.GetOrder(ctx, order.Number)
	if err != nil {
		var rateLimitErr *ErrRateLimit
		if errors.As(err, &rateLimitErr) {
			p.log.Warn("accrual rate limit",
				zap.Duration("retry_after", rateLimitErr.RetryAfter))
			p.setPause(rateLimitErr.RetryAfter)
			return
		}
		if errors.Is(err, ErrNotRegistered) {
			return
		}
		p.log.Error("get order from accrual",
			zap.String("number", order.Number), zap.Error(err))
		return
	}

	newStatus := mapAccrualStatus(result.Status)
	if newStatus == "" || newStatus == order.Status {
		return
	}

	if err := p.orderRepo.UpdateOrderStatus(ctx, order.Number, newStatus, result.Accrual); err != nil {
		p.log.Error("update order status",
			zap.String("number", order.Number), zap.Error(err))
	}
}

// mapAccrualStatus converts an accrual service status string to a domain.OrderStatus.
// Returns empty string for unknown statuses.
func mapAccrualStatus(accrualStatus string) domain.OrderStatus {
	switch accrualStatus {
	case "REGISTERED":
		return domain.OrderStatusNew
	case "PROCESSING":
		return domain.OrderStatusProcessing
	case "INVALID":
		return domain.OrderStatusInvalid
	case "PROCESSED":
		return domain.OrderStatusProcessed
	default:
		return ""
	}
}
