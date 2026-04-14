package admin

import (
	"net/http"
	"strings"
)

// tokenValidator is satisfied by Service.ValidateAdminToken.
type tokenValidator interface {
	ValidateAdminToken(tokenStr string) error
}

// Auth returns middleware that validates an admin JWT from the Authorization header.
// On failure it returns 401. Cookie-based auth is intentionally not supported;
// the admin UI uses localStorage + Authorization header via fetch.
func Auth(v tokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if err := v.ValidateAdminToken(token); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractBearer returns the token from "Authorization: Bearer <token>".
func extractBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
