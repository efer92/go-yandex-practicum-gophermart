package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"gophermart/internal/domain"
)

// OrderRepository implements repository.OrderRepo using PostgreSQL.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a new OrderRepository backed by the given pool.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

// CreateOrder inserts a new order for the given user.
// Returns domain.ErrOrderAlreadyOwnedByUser or domain.ErrOrderAlreadyOwnedByOther on conflict.
func (r *OrderRepository) CreateOrder(ctx context.Context, userID int64, number string) (*domain.Order, error) {
	var o domain.Order
	err := r.pool.QueryRow(ctx,
		`INSERT INTO orders (user_id, number) VALUES ($1, $2)
		 RETURNING id, user_id, number, status, accrual, uploaded_at`,
		userID, number,
	).Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Determine who owns the order.
			var ownerID int64
			_ = r.pool.QueryRow(ctx,
				`SELECT user_id FROM orders WHERE number = $1`, number,
			).Scan(&ownerID)
			if ownerID == userID {
				return nil, domain.ErrOrderAlreadyOwnedByUser
			}
			return nil, domain.ErrOrderAlreadyOwnedByOther
		}
		return nil, err
	}
	return &o, nil
}

// GetOrdersByUserID returns all orders for the user, newest-first.
func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at
		 FROM orders WHERE user_id = $1
		 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, rows.Err()
}

// GetPendingOrders returns up to limit orders in NEW or PROCESSING state.
func (r *OrderRepository) GetPendingOrders(ctx context.Context, limit int) ([]*domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at
		 FROM orders
		 WHERE status IN ('NEW', 'PROCESSING')
		 ORDER BY uploaded_at ASC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, rows.Err()
}

// UpdateOrderStatus updates the status and optional accrual for an order identified by number.
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, number string, status domain.OrderStatus, accrual *float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`,
		status, accrual, number,
	)
	return err
}
