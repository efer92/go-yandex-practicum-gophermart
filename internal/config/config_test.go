package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/config"
)

// unsetEnv unsets key for the duration of the test, restoring its original
// value (or absence) at cleanup. Required when we need LookupEnv to return
// ok=false — t.Setenv("KEY","") would still return ok=true.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, had := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if had {
			os.Setenv(key, prev)
		} else {
			os.Unsetenv(key)
		}
	})
}

// TestLoad_Defaults verifies that default values are applied when no configuration is provided.
func TestLoad_Defaults(t *testing.T) {
	unsetEnv(t, "RUN_ADDRESS")
	unsetEnv(t, "DATABASE_URI")
	unsetEnv(t, "ACCRUAL_SYSTEM_ADDRESS")
	unsetEnv(t, "JWT_SECRET")

	cfg := config.Load()
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost:8080", cfg.RunAddress)
	assert.NotEmpty(t, cfg.JWTSecret)
}

// TestLoad_EnvVars verifies that environment variables are applied.
func TestLoad_EnvVars(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "0.0.0.0:9090")
	t.Setenv("DATABASE_URI", "postgres://localhost/test")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
	t.Setenv("JWT_SECRET", "my-secret")

	cfg := config.Load()
	require.NotNil(t, cfg)

	assert.Equal(t, "0.0.0.0:9090", cfg.RunAddress)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURI)
	assert.Equal(t, "http://accrual:8080", cfg.AccrualSystemAddress)
	assert.Equal(t, "my-secret", cfg.JWTSecret)
}
