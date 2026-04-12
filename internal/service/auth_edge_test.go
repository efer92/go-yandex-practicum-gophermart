package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/service"
)

// TestValidateToken_EmptyString verifies an empty token is rejected.
func TestValidateToken_EmptyString(t *testing.T) {
	svc := service.NewAuthService(&mockUserRepo{}, "secret")
	_, err := svc.ValidateToken("")
	assert.Error(t, err)
}

// TestValidateToken_MalformedJWT verifies a clearly malformed JWT is rejected.
func TestValidateToken_MalformedJWT(t *testing.T) {
	svc := service.NewAuthService(&mockUserRepo{}, "secret")
	_, err := svc.ValidateToken("eyJhbGciOiJub25lIn0.e30.")
	require.Error(t, err)
}

// TestValidateToken_WrongAlgorithm verifies a token with an unexpected algorithm is rejected.
func TestValidateToken_WrongAlgorithm(t *testing.T) {
	svc := service.NewAuthService(&mockUserRepo{}, "secret")
	// This is a valid RS256 token format (but wrong alg for our HS256 service).
	_, err := svc.ValidateToken("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.invalid")
	require.Error(t, err)
}
