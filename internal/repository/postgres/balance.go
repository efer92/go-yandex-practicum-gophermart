package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophermart/internal/domain"
)

// BalanceRepository implements repository.BalanceRepo using PostgreSQL.
type BalanceRepository struct {
	pool *pgxpool.Pool
}

// NewBalanceRepository creates a new BalanceRepository backed by the given pool.
func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{pool: pool}
}

// GetBalance computes current and total withdrawn balance for a user in a single query.
func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	var b domain.Balance
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(o.accrual), 0),
			COALESCE((SELECT SUM(w.sum) FROM withdrawals w WHERE w.user_id = $1), 0)
		FROM orders o
		WHERE o.user_id = $1 AND o.status = 'PROCESSED'
	`, userID).Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		return nil, err
	}
	b.Current -= b.Withdrawn
	return &b, nil
}

// Withdraw atomically checks the balance and inserts a withdrawal record.
// Returns domain.ErrInsufficientBalance if current balance < sum.
func (r *BalanceRepository) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	var id *int64
	err := r.pool.QueryRow(ctx, `
		WITH balance AS (
			SELECT
				COALESCE(SUM(o.accrual), 0) -
				COALESCE((SELECT SUM(w.sum) FROM withdrawals w WHERE w.user_id = $1), 0) AS current
			FROM orders o
			WHERE o.user_id = $1 AND o.status = 'PROCESSED'
		)
		INSERT INTO withdrawals (user_id, order_number, sum)
		SELECT $1, $2, $3 FROM balance WHERE current >= $3
		RETURNING id
	`, userID, orderNumber, sum).Scan(&id)
	if err != nil {
		return err
	}
	if id == nil {
		return domain.ErrInsufficientBalance
	}
	return nil
}
