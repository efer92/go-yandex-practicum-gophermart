package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/repository/postgres"
)

// TestNewPool_InvalidDSN verifies that NewPool returns an error for an invalid DSN.
func TestNewPool_InvalidDSN(t *testing.T) {
	_, err := postgres.NewPool(context.Background(), "postgres://invalid-host-that-does-not-exist:5432/db?connect_timeout=1")
	assert.Error(t, err)
}

// TestNewPool_MalformedDSN verifies that NewPool returns an error for a malformed DSN.
func TestNewPool_MalformedDSN(t *testing.T) {
	_, err := postgres.NewPool(context.Background(), "not-a-dsn://!!!")
	assert.Error(t, err)
}
