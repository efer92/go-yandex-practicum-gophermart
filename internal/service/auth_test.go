package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/domain"
	"gophermart/internal/service"
)

// mockUserRepo is a test double for repository.UserRepo.
type mockUserRepo struct {
	createFn func(ctx context.Context, login, hash string) (*domain.User, error)
	getFn    func(ctx context.Context, login string) (*domain.User, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, login, hash string) (*domain.User, error) {
	return m.createFn(ctx, login, hash)
}
func (m *mockUserRepo) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return m.getFn(ctx, login)
}

// TestRegister_Success verifies a new user is created and a JWT token returned.
func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, login, hash string) (*domain.User, error) {
			return &domain.User{ID: 1, Login: login, PasswordHash: hash, CreatedAt: time.Now()}, nil
		},
	}
	svc := service.NewAuthService(repo, "test-secret")

	token, err := svc.Register(context.Background(), "alice", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token must be valid and contain the right user ID.
	userID, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "1", userID)
}

// TestRegister_Duplicate verifies ErrUserAlreadyExists is propagated.
func TestRegister_Duplicate(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _, _ string) (*domain.User, error) {
			return nil, domain.ErrUserAlreadyExists
		},
	}
	svc := service.NewAuthService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "alice", "password123")
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

// TestLogin_Success verifies a correct password yields a valid token.
func TestLogin_Success(t *testing.T) {
	// Pre-register to get a real bcrypt hash.
	repo := &mockUserRepo{}
	svc := service.NewAuthService(repo, "test-secret")

	var storedHash string
	repo.createFn = func(_ context.Context, login, hash string) (*domain.User, error) {
		storedHash = hash
		return &domain.User{ID: 42, Login: login, PasswordHash: hash}, nil
	}
	_, err := svc.Register(context.Background(), "bob", "secret")
	require.NoError(t, err)

	repo.getFn = func(_ context.Context, login string) (*domain.User, error) {
		return &domain.User{ID: 42, Login: login, PasswordHash: storedHash}, nil
	}

	token, err := svc.Login(context.Background(), "bob", "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "42", userID)
}

// TestLogin_WrongPassword verifies ErrInvalidCredentials on wrong password.
func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockUserRepo{}
	svc := service.NewAuthService(repo, "test-secret")

	var storedHash string
	repo.createFn = func(_ context.Context, login, hash string) (*domain.User, error) {
		storedHash = hash
		return &domain.User{ID: 1, Login: login, PasswordHash: hash}, nil
	}
	_, err := svc.Register(context.Background(), "carol", "rightpass")
	require.NoError(t, err)

	repo.getFn = func(_ context.Context, login string) (*domain.User, error) {
		return &domain.User{ID: 1, Login: login, PasswordHash: storedHash}, nil
	}

	_, err = svc.Login(context.Background(), "carol", "wrongpass")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

// TestLogin_UnknownUser verifies ErrInvalidCredentials when user doesn't exist.
func TestLogin_UnknownUser(t *testing.T) {
	repo := &mockUserRepo{
		getFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, nil // not found
		},
	}
	svc := service.NewAuthService(repo, "test-secret")

	_, err := svc.Login(context.Background(), "ghost", "pass")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

// TestValidateToken_InvalidToken verifies that a tampered token is rejected.
func TestValidateToken_InvalidToken(t *testing.T) {
	repo := &mockUserRepo{}
	svc := service.NewAuthService(repo, "test-secret")

	_, err := svc.ValidateToken("not.a.valid.jwt")
	assert.Error(t, err)
}

// TestValidateToken_WrongSecret verifies that a token signed with a different secret is rejected.
func TestValidateToken_WrongSecret(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, login, hash string) (*domain.User, error) {
			return &domain.User{ID: 1, Login: login, PasswordHash: hash}, nil
		},
	}
	svc1 := service.NewAuthService(repo, "secret-a")
	svc2 := service.NewAuthService(repo, "secret-b")

	token, err := svc1.Register(context.Background(), "user", "pass")
	require.NoError(t, err)

	_, err = svc2.ValidateToken(token)
	assert.Error(t, err)
}

// TestLogin_RepoError verifies that repository errors are propagated.
func TestLogin_RepoError(t *testing.T) {
	repo := &mockUserRepo{
		getFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, errors.New("db error")
		},
	}
	svc := service.NewAuthService(repo, "test-secret")
	_, err := svc.Login(context.Background(), "user", "pass")
	assert.Error(t, err)
}
