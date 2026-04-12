package domain

import "errors"

// Sentinel errors used throughout the service layer.
// Handlers use errors.Is to map these to HTTP status codes.
var (
	// ErrUserAlreadyExists is returned when trying to register a login that already exists.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidCredentials is returned when login/password do not match.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrOrderAlreadyOwnedByUser is returned when the same user submits an already-tracked order.
	ErrOrderAlreadyOwnedByUser = errors.New("order already submitted by this user")

	// ErrOrderAlreadyOwnedByOther is returned when a different user already owns the order.
	ErrOrderAlreadyOwnedByOther = errors.New("order already submitted by another user")

	// ErrInvalidOrderNumber is returned when the order number fails the Luhn check.
	ErrInvalidOrderNumber = errors.New("invalid order number")

	// ErrInsufficientBalance is returned when a withdrawal exceeds the current balance.
	ErrInsufficientBalance = errors.New("insufficient balance")
)
