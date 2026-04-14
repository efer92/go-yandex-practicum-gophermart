package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gophermart/internal/domain"
	"gophermart/internal/middleware"
)

// WithdrawalProvider is the subset of WithdrawalService used by withdrawal handlers.
type WithdrawalProvider interface {
	// ListWithdrawals returns all withdrawals for the user.
	ListWithdrawals(ctx context.Context, userID int64) ([]*domain.Withdrawal, error)
}

// WithdrawalHandler handles withdrawal history requests.
type WithdrawalHandler struct {
	withdrawals WithdrawalProvider
}

// NewWithdrawalHandler creates a new WithdrawalHandler.
func NewWithdrawalHandler(withdrawals WithdrawalProvider) *WithdrawalHandler {
	return &WithdrawalHandler{withdrawals: withdrawals}
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// ListWithdrawals handles GET /api/user/withdrawals.
func (h *WithdrawalHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(middleware.UserIDFromCtx(r.Context()), 10, 64)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	list, err := h.withdrawals.ListWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]withdrawalResponse, len(list))
	for i, wd := range list {
		resp[i] = withdrawalResponse{
			Order:       wd.OrderNumber,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
