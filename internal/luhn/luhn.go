// Package luhn provides the Luhn algorithm for order number validation.
package luhn

// Valid reports whether the given number string passes the Luhn check.
// The input must consist of digits only; any other character returns false.
// An empty string or a single-digit string returns false.
func Valid(number string) bool {
	if len(number) < 2 {
		return false
	}

	sum := 0
	nDigits := len(number)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		ch := number[i]
		if ch < '0' || ch > '9' {
			return false
		}
		digit := int(ch - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
