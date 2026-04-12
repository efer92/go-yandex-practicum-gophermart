package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"gophermart/internal/domain"
)

// WithdrawalRepository implements repository.WithdrawalRepo using PostgreSQL.
type WithdrawalRepository struct {
	pool *pgxpool.Pool
}

// NewWithdrawalRepository creates a new WithdrawalRepository backed by the given pool.
func NewWithdrawalRepository(pool *pgxpool.Pool) *WithdrawalRepository {
	return &WithdrawalRepository{pool: pool}
}

// GetWithdrawalsByUserID returns all withdrawals for the user, newest-first.
func (r *WithdrawalRepository) GetWithdrawalsByUserID(ctx context.Context, userID int64) ([]*domain.Withdrawal, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT order_number, sum, processed_at
		 FROM withdrawals WHERE user_id = $1
		 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*domain.Withdrawal
	for rows.Next() {
		var w domain.Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, rows.Err()
}
