// Package domain contains the core domain types for the Gophermart loyalty system.
package domain

import "time"

// OrderStatus represents the processing status of a loyalty order.
type OrderStatus string

const (
	// OrderStatusNew means the order has been uploaded but not yet processed.
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusProcessing means the accrual calculation is in progress.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusInvalid means the accrual system rejected the order.
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusProcessed means the accrual calculation is complete.
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// User represents a registered user of the loyalty system.
type User struct {
	// ID is the unique database identifier.
	ID int64
	// Login is the unique username.
	Login string
	// PasswordHash is the bcrypt hash of the user's password.
	PasswordHash string
	// CreatedAt is the time of registration.
	CreatedAt time.Time
}

// Order represents a loyalty order submitted by a user.
type Order struct {
	// ID is the unique database identifier.
	ID int64
	// UserID is the ID of the user who submitted this order.
	UserID int64
	// Number is the unique order number (validated by Luhn algorithm).
	Number string
	// Status is the current processing status.
	Status OrderStatus
	// Accrual is the points awarded for this order (nil until PROCESSED).
	Accrual *float64
	// UploadedAt is the time the order was submitted.
	UploadedAt time.Time
}

// Balance represents a user's current loyalty points balance.
type Balance struct {
	// Current is the available balance (accrued minus withdrawn).
	Current float64 `json:"current"`
	// Withdrawn is the total amount ever withdrawn.
	Withdrawn float64 `json:"withdrawn"`
}

// Withdrawal represents a single withdrawal of loyalty points.
type Withdrawal struct {
	// OrderNumber is the order number used for this withdrawal.
	OrderNumber string `json:"order"`
	// Sum is the amount of points withdrawn.
	Sum float64 `json:"sum"`
	// ProcessedAt is the time of the withdrawal.
	ProcessedAt time.Time `json:"processed_at"`
}
