package domain_test

import (
	"errors"
	"testing"

	"gophermart/internal/domain"
)

// TestSentinelErrors verifies that all sentinel errors are distinct and identifiable via errors.Is.
func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		domain.ErrUserAlreadyExists,
		domain.ErrInvalidCredentials,
		domain.ErrOrderAlreadyOwnedByUser,
		domain.ErrOrderAlreadyOwnedByOther,
		domain.ErrInvalidOrderNumber,
		domain.ErrInsufficientBalance,
	}

	for i, err := range sentinels {
		if err == nil {
			t.Errorf("sentinel[%d] is nil", i)
			continue
		}
		if !errors.Is(err, err) {
			t.Errorf("errors.Is(%v, %v) = false, want true", err, err)
		}
		// Ensure sentinels are distinct from each other.
		for j, other := range sentinels {
			if i != j && errors.Is(err, other) {
				t.Errorf("sentinel[%d] == sentinel[%d]: %v", i, j, err)
			}
		}
	}
}
