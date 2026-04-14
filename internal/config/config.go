// Package config handles configuration for the Gophermart service.
// Configuration is read from command-line flags and environment variables.
// Flags take precedence over environment variables.
package config

import (
	"flag"
	"os"
)

// Config holds all runtime configuration for the service.
type Config struct {
	// RunAddress is the address and port for the HTTP server (e.g. "localhost:8080").
	RunAddress string
	// DatabaseURI is the PostgreSQL connection string.
	DatabaseURI string
	// AccrualSystemAddress is the base URL of the external accrual service.
	AccrualSystemAddress string
	// JWTSecret is the HMAC secret used for signing JWT tokens.
	JWTSecret string
	// AdminLogin is the username for the admin panel (env: ADMIN_LOGIN, default: "admin").
	AdminLogin string
	// AdminPassword is the password for the admin panel (env: ADMIN_PASSWORD, default: "admin").
	AdminPassword string
}

// Load parses configuration from flags and environment variables.
// Flag values override environment variables when both are set.
// LookupEnv is used to distinguish between an unset variable and one explicitly
// set to an empty string — the latter is honoured as a valid value.
func Load() *Config {
	cfg := &Config{}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", "", "server listen address (host:port)")
	fs.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL connection string")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system base URL")
	// Ignore errors from unknown flags during tests.
	fs.Parse(os.Args[1:])

	if cfg.RunAddress == "" {
		cfg.RunAddress = lookupEnvOrDefault("RUN_ADDRESS", "localhost:8080")
	}
	if cfg.DatabaseURI == "" {
		if v, ok := os.LookupEnv("DATABASE_URI"); ok {
			cfg.DatabaseURI = v
		}
	}
	if cfg.AccrualSystemAddress == "" {
		if v, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
			cfg.AccrualSystemAddress = v
		}
	}

	cfg.JWTSecret = lookupEnvOrDefault("JWT_SECRET", "gophermart-default-secret-change-in-prod")
	cfg.AdminLogin = lookupEnvOrDefault("ADMIN_LOGIN", "admin")
	cfg.AdminPassword = lookupEnvOrDefault("ADMIN_PASSWORD", "admin")

	return cfg
}

// lookupEnvOrDefault returns the environment variable value if it is set (even
// if empty), or defaultVal if the variable is not defined at all.
func lookupEnvOrDefault(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
