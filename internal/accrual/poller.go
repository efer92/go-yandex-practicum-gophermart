package accrual

import (
	"context"
	"errors"
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
	client    *Client
	orderRepo repository.OrderRepo
	log       *zap.Logger
}

// NewPoller creates a new Poller.
func NewPoller(client *Client, orderRepo repository.OrderRepo, log *zap.Logger) *Poller {
	return &Poller{
		client:    client,
		orderRepo: orderRepo,
		log:       log,
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
						if pause := p.processOrder(ctx, order); pause > 0 {
							ticker.Reset(pause)
						}
					}()
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// processOrder fetches accrual info for one order and updates the DB.
// Returns a non-zero duration if the poller should back off (rate limit).
func (p *Poller) processOrder(ctx context.Context, order *domain.Order) time.Duration {
	result, err := p.client.GetOrder(ctx, order.Number)
	if err != nil {
		var rateLimitErr *ErrRateLimit
		if errors.As(err, &rateLimitErr) {
			p.log.Warn("accrual rate limit", zap.Duration("retry_after", rateLimitErr.RetryAfter))
			return rateLimitErr.RetryAfter
		}
		if errors.Is(err, ErrNotRegistered) {
			// Order not yet in accrual system; leave as NEW.
			return 0
		}
		p.log.Error("get order from accrual", zap.String("number", order.Number), zap.Error(err))
		return 0
	}

	newStatus := mapAccrualStatus(result.Status)
	if newStatus == "" || newStatus == order.Status {
		return 0
	}

	if err := p.orderRepo.UpdateOrderStatus(ctx, order.Number, newStatus, result.Accrual); err != nil {
		p.log.Error("update order status", zap.String("number", order.Number), zap.Error(err))
	}
	return 0
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
