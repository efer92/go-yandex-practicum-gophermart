package luhn_test

import (
	"testing"

	"gophermart/internal/luhn"
)

// TestValid verifies the Luhn validation against known valid and invalid numbers.
func TestValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		// Valid numbers from the specification examples.
		{name: "spec example 1", number: "9278923470", want: true},
		{name: "spec example 2", number: "12345678903", want: true},
		{name: "spec example withdraw", number: "2377225624", want: true},
		{name: "spec example 3", number: "346436439", want: true},

		// Classic Luhn-valid numbers.
		{name: "visa test", number: "4532015112830366", want: true},
		{name: "mastercard test", number: "5425233430109903", want: true},
		{name: "amex test", number: "378282246310005", want: true},
		{name: "two digit valid", number: "18", want: true},

		// Invalid numbers.
		{name: "all zeros one digit", number: "0", want: false},
		{name: "empty string", number: "", want: false},
		{name: "single digit", number: "1", want: false},
		{name: "wrong check digit", number: "9278923471", want: false},
		{name: "wrong check digit 2", number: "12345678900", want: false},
		{name: "non-digit char", number: "1234a5678", want: false},
		{name: "non-digit space", number: "1234 5678", want: false},
		{name: "letters only", number: "abcdef", want: false},
		{name: "two digit 00", number: "00", want: true},
		{name: "all nines", number: "99", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := luhn.Valid(tt.number)
			if got != tt.want {
				t.Errorf("Valid(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}
