package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler handles health check requests
type Handler struct {
	ready bool
}

// NewHandler creates a new health handler
func NewHandler() *Handler {
	return &Handler{ready: true}
}

// RegisterRoutes registers health check routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health/live", h.Liveness)
	r.Get("/health/ready", h.Readiness)
}

// Liveness handles GET /health/live
// Returns 200 if the service is alive
func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness handles GET /health/ready
// Returns 200 if the service is ready to accept traffic
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	if !h.ready {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
