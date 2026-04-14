// Package handler provides HTTP handlers for the Gophermart loyalty system.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gophermart/internal/domain"
	"gophermart/internal/middleware"
)

// AuthProvider is the subset of AuthService used by auth handlers.
type AuthProvider interface {
	// Register creates a new user and returns a signed JWT token.
	Register(ctx context.Context, login, password string) (string, error)
	// Login authenticates a user and returns a signed JWT token.
	Login(ctx context.Context, login, password string) (string, error)
}

// AuthHandler handles user registration and login.
type AuthHandler struct {
	auth AuthProvider
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth AuthProvider) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register handles POST /api/user/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	middleware.AddLogFields(r, zap.String("login", req.Login))
	token, err := h.auth.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			http.Error(w, "login already taken", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	setAuthToken(w, token)
	w.WriteHeader(http.StatusOK)
}

// Login handles POST /api/user/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	middleware.AddLogFields(r, zap.String("login", req.Login))
	token, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	setAuthToken(w, token)
	w.WriteHeader(http.StatusOK)
}

// setAuthToken writes the JWT token to both the Authorization header and a cookie.
func setAuthToken(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", "Bearer "+token)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
	})
}
