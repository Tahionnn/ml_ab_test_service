package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
	mw "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport/middleware"
	"go.uber.org/zap"
)

type HealthHandler struct {
	svc *gateway.Service
	lg  *zap.Logger
}

func NewHealthHandler(svc *gateway.Service, lg *zap.Logger) *HealthHandler {
	return &HealthHandler{svc: svc, lg: lg}
}

func (h *HealthHandler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	r.Get("/health", h.Health)
	r.Group(func(r chi.Router) {
		r.Use(auth, mw.RequireRole(user.Admin))
		r.Get("/health/deep", h.DeepHealth)
	})
}

// Health godoc
// @Summary      Basic health check (always returns 200)
// @Tags         system
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeepHealth godoc
// @Summary      Deep health check — pings all downstream services (admin only)
// @Description  Checks connectivity to user service, experiment manager, model registry.
// @Tags         system
// @Security     BearerAuth
// @Success      200  {array}   gateway.HealthStatus
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Admin role required"
// @Router       /health/deep [get]
func (h *HealthHandler) DeepHealth(w http.ResponseWriter, r *http.Request) {
	statuses := h.svc.CheckHealth(r.Context())
	writeJSON(w, http.StatusOK, statuses)
}
