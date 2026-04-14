package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/middleware"
)

// mockValidator is a test double for middleware.TokenValidator.
type mockValidator struct {
	validateFn func(token string) (string, error)
}

func (m *mockValidator) ValidateToken(token string) (string, error) {
	return m.validateFn(token)
}

// nextHandler is a simple handler that captures the user ID from context.
func nextHandler(t *testing.T, expectedUserID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := middleware.UserIDFromCtx(r.Context())
		assert.Equal(t, expectedUserID, uid)
		w.WriteHeader(http.StatusOK)
	})
}

// TestAuthMiddleware_ValidHeaderToken verifies the middleware passes with a Bearer token.
func TestAuthMiddleware_ValidHeaderToken(t *testing.T) {
	validator := &mockValidator{
		validateFn: func(token string) (string, error) {
			assert.Equal(t, "valid-token", token)
			return "42", nil
		},
	}

	handler := middleware.Auth(validator)(nextHandler(t, "42"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestAuthMiddleware_ValidCookieToken verifies the middleware passes with a cookie token.
func TestAuthMiddleware_ValidCookieToken(t *testing.T) {
	validator := &mockValidator{
		validateFn: func(token string) (string, error) { return "7", nil },
	}

	handler := middleware.Auth(validator)(nextHandler(t, "7"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "cookie-token"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestAuthMiddleware_NoToken verifies 401 when no token is provided.
func TestAuthMiddleware_NoToken(t *testing.T) {
	validator := &mockValidator{}
	handler := middleware.Auth(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestAuthMiddleware_InvalidToken verifies 401 when token validation fails.
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	validator := &mockValidator{
		validateFn: func(_ string) (string, error) { return "", fmt.Errorf("bad token") },
	}
	handler := middleware.Auth(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestWithUserID verifies that WithUserID and UserIDFromCtx round-trip correctly.
func TestWithUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.WithUserID(req.Context(), "99")
	assert.Equal(t, "99", middleware.UserIDFromCtx(ctx))
}

// TestUserIDFromCtx_Missing verifies empty string is returned when no user ID is in context.
func TestUserIDFromCtx_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", middleware.UserIDFromCtx(req.Context()))
}
