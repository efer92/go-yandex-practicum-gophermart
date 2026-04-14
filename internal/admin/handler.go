package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

// dataProvider is implemented by PostgresRepo.
type dataProvider interface {
	GetStats(ctx context.Context) (*Stats, error)
	GetUsers(ctx context.Context) ([]*UserRow, error)
	GetOrders(ctx context.Context) ([]*OrderRow, error)
	GetWithdrawals(ctx context.Context) ([]*WithdrawalRow, error)
}

// loginProvider is implemented by Service.
type loginProvider interface {
	Login(login, password string) (string, error)
}

// Handler handles all admin HTTP API requests.
type Handler struct {
	svc  loginProvider
	repo dataProvider
	log  *zap.Logger
}

// NewHandler creates a new admin Handler.
func NewHandler(svc loginProvider, repo dataProvider, log *zap.Logger) *Handler {
	return &Handler{svc: svc, repo: repo, log: log}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Login handles POST /admin/api/login.
// On success returns {"token":"<jwt>"} with 200.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Login == "" || req.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidAdminCredentials) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.log.Error("admin login", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// GetStats handles GET /admin/api/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetStats(r.Context())
	if err != nil {
		h.log.Error("admin get stats", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

// GetUsers handles GET /admin/api/users.
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetUsers(r.Context())
	if err != nil {
		h.log.Error("admin get users", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []*UserRow{}
	}
	writeJSON(w, users)
}

// GetOrders handles GET /admin/api/orders.
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.repo.GetOrders(r.Context())
	if err != nil {
		h.log.Error("admin get orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if orders == nil {
		orders = []*OrderRow{}
	}
	writeJSON(w, orders)
}

// GetWithdrawals handles GET /admin/api/withdrawals.
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	withdrawals, err := h.repo.GetWithdrawals(r.Context())
	if err != nil {
		h.log.Error("admin get withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if withdrawals == nil {
		withdrawals = []*WithdrawalRow{}
	}
	writeJSON(w, withdrawals)
}

// writeJSON encodes v as JSON and sets the Content-Type header.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
