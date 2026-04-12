package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/config"
)

// TestLoad_Defaults verifies that default values are applied when no configuration is provided.
func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that might be set.
	os.Unsetenv("RUN_ADDRESS")
	os.Unsetenv("DATABASE_URI")
	os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
	os.Unsetenv("JWT_SECRET")

	cfg := config.Load()
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost:8080", cfg.RunAddress)
	assert.NotEmpty(t, cfg.JWTSecret)
}

// TestLoad_EnvVars verifies that environment variables are applied.
func TestLoad_EnvVars(t *testing.T) {
	os.Setenv("RUN_ADDRESS", "0.0.0.0:9090")
	os.Setenv("DATABASE_URI", "postgres://localhost/test")
	os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
	os.Setenv("JWT_SECRET", "my-secret")
	defer func() {
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := config.Load()
	require.NotNil(t, cfg)

	assert.Equal(t, "0.0.0.0:9090", cfg.RunAddress)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURI)
	assert.Equal(t, "http://accrual:8080", cfg.AccrualSystemAddress)
	assert.Equal(t, "my-secret", cfg.JWTSecret)
}
