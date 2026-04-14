package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"gophermart/internal/domain"
)

// UserRepository implements repository.UserRepo using PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository backed by the given pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser inserts a new user record. Returns domain.ErrUserAlreadyExists on unique constraint violation.
func (r *UserRepository) CreateUser(ctx context.Context, login, passwordHash string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2)
		 RETURNING id, login, password_hash, created_at`,
		login, passwordHash,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, err
	}
	return &u, nil
}

// GetUserByLogin retrieves a user by login. Returns nil, nil if not found.
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at FROM users WHERE login = $1`,
		login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
