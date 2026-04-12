package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/handler"
)

// TestRegisterHandler_InternalError verifies 500 when the auth service fails unexpectedly.
func TestRegisterHandler_InternalError(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{
		registerFn: func(_ context.Context, _, _ string) (string, error) {
			return "", assert.AnError
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", authBody("alice", "pass"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestLoginHandler_BadRequest verifies 400 on malformed JSON body.
func TestLoginHandler_BadRequest(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", authBody("", ""))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
