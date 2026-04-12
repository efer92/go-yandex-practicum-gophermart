package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"gophermart/internal/domain"
	"gophermart/internal/middleware"
)

// OrderProvider is the subset of OrderService used by order handlers.
type OrderProvider interface {
	// SubmitOrder validates and records a new order.
	SubmitOrder(ctx context.Context, userID int64, number string) (*domain.Order, error)
	// ListOrders returns all orders for the user.
	ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error)
}

// OrderHandler handles order submission and listing.
type OrderHandler struct {
	orders OrderProvider
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orders OrderProvider) *OrderHandler {
	return &OrderHandler{orders: orders}
}

// orderResponse is the JSON shape for a single order in list responses.
type orderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

// UploadOrder handles POST /api/user/orders.
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	middleware.AddLogFields(r, zap.String("order", number))
	userID := middleware.UserIDFromCtx(r.Context())
	_, err = h.orders.SubmitOrder(r.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOrderNumber):
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrOrderAlreadyOwnedByUser):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, domain.ErrOrderAlreadyOwnedByOther):
			http.Error(w, "order owned by another user", http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// ListOrders handles GET /api/user/orders.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	orders, err := h.orders.ListOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]orderResponse, len(orders))
	for i, o := range orders {
		resp[i] = orderResponse{
			Number:     o.Number,
			Status:     string(o.Status),
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
