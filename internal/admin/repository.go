// Package admin provides the admin panel backend for the Gophermart loyalty system.
package admin

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Stats contains aggregate statistics for the admin dashboard.
type Stats struct {
	// TotalUsers is the number of registered users.
	TotalUsers int64 `json:"total_users"`
	// CountNew is the number of orders in NEW status.
	CountNew int64 `json:"count_new"`
	// CountProcessing is the number of orders in PROCESSING status.
	CountProcessing int64 `json:"count_processing"`
	// CountInvalid is the number of orders in INVALID status.
	CountInvalid int64 `json:"count_invalid"`
	// CountProcessed is the number of orders in PROCESSED status.
	CountProcessed int64 `json:"count_processed"`
	// TotalAccrual is the total loyalty points accrued across all users.
	TotalAccrual float64 `json:"total_accrual"`
	// TotalWithdrawn is the total loyalty points withdrawn across all users.
	TotalWithdrawn float64 `json:"total_withdrawn"`
}

// UserRow is the admin projection of a user with computed balance.
type UserRow struct {
	// ID is the user's database identifier.
	ID int64 `json:"id"`
	// Login is the user's login name.
	Login string `json:"login"`
	// CreatedAt is the registration timestamp.
	CreatedAt time.Time `json:"created_at"`
	// Balance is the current available balance (accrued minus withdrawn).
	Balance float64 `json:"balance"`
	// TotalOrders is the number of orders submitted by this user.
	TotalOrders int64 `json:"total_orders"`
}

// OrderRow is the admin projection of an order with the owner's login.
type OrderRow struct {
	// ID is the order's database identifier.
	ID int64 `json:"id"`
	// UserLogin is the login of the user who submitted this order.
	UserLogin string `json:"user_login"`
	// Number is the order number.
	Number string `json:"number"`
	// Status is the current processing status.
	Status string `json:"status"`
	// Accrual is the points awarded (nil if not yet PROCESSED).
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt is the submission timestamp.
	UploadedAt time.Time `json:"uploaded_at"`
}

// WithdrawalRow is the admin projection of a withdrawal with the owner's login.
type WithdrawalRow struct {
	// ID is the withdrawal's database identifier.
	ID int64 `json:"id"`
	// UserLogin is the login of the user who made this withdrawal.
	UserLogin string `json:"user_login"`
	// OrderNumber is the order number used for the withdrawal.
	OrderNumber string `json:"order_number"`
	// Sum is the number of points withdrawn.
	Sum float64 `json:"sum"`
	// ProcessedAt is the withdrawal timestamp.
	ProcessedAt time.Time `json:"processed_at"`
}

// Repo provides read-only admin queries spanning all users.
type Repo interface {
	// GetStats returns aggregate statistics for the dashboard.
	GetStats(ctx context.Context) (*Stats, error)
	// GetUsers returns all users with computed balances, newest-first.
	GetUsers(ctx context.Context) ([]*UserRow, error)
	// GetOrders returns all orders with user logins, newest-first.
	GetOrders(ctx context.Context) ([]*OrderRow, error)
	// GetWithdrawals returns all withdrawals with user logins, newest-first.
	GetWithdrawals(ctx context.Context) ([]*WithdrawalRow, error)
}

// PostgresRepo implements Repo using PostgreSQL.
type PostgresRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresRepo creates a new PostgresRepo backed by the given pool.
func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

// GetStats returns aggregate statistics in a single query.
func (r *PostgresRepo) GetStats(ctx context.Context) (*Stats, error) {
	var s Stats
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users)                                              AS total_users,
			COUNT(*) FILTER (WHERE status = 'NEW')                                    AS count_new,
			COUNT(*) FILTER (WHERE status = 'PROCESSING')                             AS count_processing,
			COUNT(*) FILTER (WHERE status = 'INVALID')                                AS count_invalid,
			COUNT(*) FILTER (WHERE status = 'PROCESSED')                              AS count_processed,
			COALESCE(SUM(accrual) FILTER (WHERE status = 'PROCESSED'), 0)             AS total_accrual,
			(SELECT COALESCE(SUM(sum), 0) FROM withdrawals)                           AS total_withdrawn
		FROM orders
	`).Scan(
		&s.TotalUsers,
		&s.CountNew,
		&s.CountProcessing,
		&s.CountInvalid,
		&s.CountProcessed,
		&s.TotalAccrual,
		&s.TotalWithdrawn,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetUsers returns all users with computed balances, ordered by registration date descending.
func (r *PostgresRepo) GetUsers(ctx context.Context) ([]*UserRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			u.id,
			u.login,
			u.created_at,
			COALESCE(SUM(o.accrual) FILTER (WHERE o.status = 'PROCESSED'), 0)
				- COALESCE((SELECT SUM(w.sum) FROM withdrawals w WHERE w.user_id = u.id), 0) AS balance,
			COUNT(o.id) AS total_orders
		FROM users u
		LEFT JOIN orders o ON o.user_id = u.id
		GROUP BY u.id, u.login, u.created_at
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*UserRow
	for rows.Next() {
		var row UserRow
		if err := rows.Scan(&row.ID, &row.Login, &row.CreatedAt, &row.Balance, &row.TotalOrders); err != nil {
			return nil, err
		}
		result = append(result, &row)
	}
	return result, rows.Err()
}

// GetOrders returns all orders with user logins, ordered by upload date descending.
func (r *PostgresRepo) GetOrders(ctx context.Context) ([]*OrderRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, u.login, o.number, o.status, o.accrual, o.uploaded_at
		FROM orders o
		JOIN users u ON u.id = o.user_id
		ORDER BY o.uploaded_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*OrderRow
	for rows.Next() {
		var row OrderRow
		if err := rows.Scan(&row.ID, &row.UserLogin, &row.Number, &row.Status, &row.Accrual, &row.UploadedAt); err != nil {
			return nil, err
		}
		result = append(result, &row)
	}
	return result, rows.Err()
}

// GetWithdrawals returns all withdrawals with user logins, ordered by date descending.
func (r *PostgresRepo) GetWithdrawals(ctx context.Context) ([]*WithdrawalRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT w.id, u.login, w.order_number, w.sum, w.processed_at
		FROM withdrawals w
		JOIN users u ON u.id = w.user_id
		ORDER BY w.processed_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*WithdrawalRow
	for rows.Next() {
		var row WithdrawalRow
		if err := rows.Scan(&row.ID, &row.UserLogin, &row.OrderNumber, &row.Sum, &row.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, &row)
	}
	return result, rows.Err()
}
