package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"gophermart/internal/domain"
	"gophermart/internal/middleware"
)

// BalanceProvider is the subset of BalanceService used by balance handlers.
type BalanceProvider interface {
	// GetBalance returns the current balance for the user.
	GetBalance(ctx context.Context, userID int64) (*domain.Balance, error)
	// Withdraw deducts sum points from the user's balance.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error
}

// BalanceHandler handles balance retrieval and withdrawal requests.
type BalanceHandler struct {
	balance BalanceProvider
}

// NewBalanceHandler creates a new BalanceHandler.
func NewBalanceHandler(balance BalanceProvider) *BalanceHandler {
	return &BalanceHandler{balance: balance}
}

// GetBalance handles GET /api/user/balance.
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	bal, err := h.balance.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bal)
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Withdraw handles POST /api/user/balance/withdraw.
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Order == "" || req.Sum <= 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	userID := middleware.UserIDFromCtx(r.Context())
	err := h.balance.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOrderNumber):
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrInsufficientBalance):
			http.Error(w, "insufficient balance", http.StatusPaymentRequired)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
