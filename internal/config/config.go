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
func Load() *Config {
	cfg := &Config{}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", "", "server listen address (host:port)")
	fs.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL connection string")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system base URL")
	// Ignore errors from unknown flags during tests.
	fs.Parse(os.Args[1:])

	if cfg.RunAddress == "" {
		cfg.RunAddress = os.Getenv("RUN_ADDRESS")
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = os.Getenv("DATABASE_URI")
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "gophermart-default-secret-change-in-prod"
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = "localhost:8080"
	}

	cfg.AdminLogin = os.Getenv("ADMIN_LOGIN")
	if cfg.AdminLogin == "" {
		cfg.AdminLogin = "admin"
	}
	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = "admin"
	}

	return cfg
}
