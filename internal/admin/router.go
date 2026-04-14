package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gophermart/web"
)

// NewRouter assembles the admin chi subrouter.
// It is mounted at /admin by the top-level mux, so all paths here are relative:
//   - GET  /          → serves admin.html
//   - POST /api/login → public login endpoint
//   - GET  /api/*     → protected by admin JWT middleware
func NewRouter(h *Handler, svc *Service) http.Handler {
	r := chi.NewRouter()

	// Serve the embedded admin SPA.
	r.Get("/", serveUI)
	r.Get("/index.html", serveUI)

	// Public: login.
	r.Post("/api/login", h.Login)

	// Protected admin API routes.
	r.Group(func(r chi.Router) {
		r.Use(Auth(svc))
		r.Get("/api/stats", h.GetStats)
		r.Get("/api/users", h.GetUsers)
		r.Get("/api/orders", h.GetOrders)
		r.Get("/api/withdrawals", h.GetWithdrawals)
	})

	return r
}

// serveUI writes the embedded admin.html to the response.
func serveUI(w http.ResponseWriter, r *http.Request) {
	data, err := web.FS.ReadFile("admin.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}
