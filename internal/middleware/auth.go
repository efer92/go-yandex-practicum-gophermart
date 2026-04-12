// Package middleware provides HTTP middleware for the Gophermart service.
package middleware

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys in this package.
type contextKey int

const userIDKey contextKey = 0

// TokenValidator is the subset of AuthService used by the auth middleware.
type TokenValidator interface {
	// ValidateToken parses and validates a JWT token, returning the user ID.
	ValidateToken(tokenStr string) (int64, error)
}

// Auth returns middleware that extracts and validates a JWT token from the
// Authorization header (Bearer scheme) or the "token" cookie.
// On success, the user ID is stored in the request context via UserIDFromCtx.
// On failure, 401 Unauthorized is returned.
func Auth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := validator.ValidateToken(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromCtx retrieves the authenticated user ID from the context.
// Returns 0 if no user ID is present (i.e., on unauthenticated routes).
func UserIDFromCtx(ctx context.Context) int64 {
	id, _ := ctx.Value(userIDKey).(int64)
	return id
}

// WithUserID returns a context with the given user ID injected.
// This is intended for use in tests and server-side helpers.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}


// extractToken returns the raw JWT string from the Authorization header or cookie.
func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}
	return ""
}
