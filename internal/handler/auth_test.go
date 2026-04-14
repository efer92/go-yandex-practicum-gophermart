package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gophermart/internal/domain"
	"gophermart/internal/handler"
)

// mockAuthProvider is a test double for handler.AuthProvider.
type mockAuthProvider struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
	loginFn    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockAuthProvider) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}
func (m *mockAuthProvider) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

func authBody(login, password string) *bytes.Buffer {
	b, _ := json.Marshal(map[string]string{"login": login, "password": password})
	return bytes.NewBuffer(b)
}

// TestRegisterHandler_Success verifies 200 and token on successful registration.
func TestRegisterHandler_Success(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{
		registerFn: func(_ context.Context, _, _ string) (string, error) { return "token123", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", authBody("alice", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Authorization"), "token123")
}

// TestRegisterHandler_Conflict verifies 409 when login is taken.
func TestRegisterHandler_Conflict(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{
		registerFn: func(_ context.Context, _, _ string) (string, error) {
			return "", domain.ErrUserAlreadyExists
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", authBody("alice", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

// TestRegisterHandler_BadRequest verifies 400 on malformed JSON.
func TestRegisterHandler_BadRequest(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestRegisterHandler_EmptyLogin verifies 400 when login is empty.
func TestRegisterHandler_EmptyLogin(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", authBody("", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestLoginHandler_Success verifies 200 on valid credentials.
func TestLoginHandler_Success(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{
		loginFn: func(_ context.Context, _, _ string) (string, error) { return "jwt-token", nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", authBody("alice", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestLoginHandler_Unauthorized verifies 401 on wrong credentials.
func TestLoginHandler_Unauthorized(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{
		loginFn: func(_ context.Context, _, _ string) (string, error) {
			return "", domain.ErrInvalidCredentials
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", authBody("alice", "wrong"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestLoginHandler_InternalError verifies 500 on unexpected error.
func TestLoginHandler_InternalError(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{
		loginFn: func(_ context.Context, _, _ string) (string, error) {
			return "", assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", authBody("alice", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
