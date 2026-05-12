package httptransport

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/experiment"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
	mw "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport/middleware"
	"go.uber.org/zap"
)

type ExperimentHandler struct {
	svc *gateway.Service
	lg  *zap.Logger
}

func NewExperimentHandler(svc *gateway.Service, lg *zap.Logger) *ExperimentHandler {
	return &ExperimentHandler{svc, lg}
}

func (h *ExperimentHandler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	scientist := mw.RequireRole(user.Scientist, user.Admin)
	admin := mw.RequireRole(user.Admin)

	r.Group(func(r chi.Router) {
		r.Use(auth)

		// Experiments
		r.Get("/experiments/{id}", h.GetExperiment) // viewer+
		r.With(scientist).Post("/experiments", h.Create)
		r.With(scientist).Put("/experiments/{id}", h.Update)
		r.With(admin).Delete("/experiments/{id}", h.Delete)

		// Lifecycle (scientist+)
		r.With(scientist).Post("/experiments/{id}/start", h.Start)
		r.With(scientist).Post("/experiments/{id}/stop", h.Stop)
		r.With(scientist).Post("/experiments/{id}/resume", h.Resume)
		r.With(scientist).Post("/experiments/{id}/finish", h.Finish)
		r.With(scientist).Post("/experiments/{id}/restore", h.Restore)

		// Variants
		r.Get("/experiments/{id}/variants", h.ListVariants)
		r.With(scientist).Post("/experiments/{id}/variants", h.AddVariant)
		r.Get("/variants/{id}", h.GetVariant)
		r.With(scientist).Put("/variants/{id}", h.UpdateVariant)
		r.With(scientist).Delete("/variants/{id}", h.DeleteVariant)

		// Metrics
		r.Get("/metrics", h.ListMetrics)
		r.Get("/metrics/{id}", h.GetMetric)
		r.With(scientist).Post("/metrics", h.CreateMetric)
		r.With(scientist).Put("/metrics/{id}", h.UpdateMetric)
		r.With(admin).Delete("/metrics/{id}", h.DeleteMetric)

		r.Get("/experiments/{id}/metrics", h.ListExperimentMetrics)
		r.With(scientist).Post("/experiments/{id}/metrics", h.AttachMetric)
		r.With(scientist).Delete("/experiments/{id}/metrics/{metricId}", h.DetachMetric)
	})
}

// GetExperiment godoc
// @Summary      Get experiment by ID
// @Tags         experiments
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Experiment ID"
// @Success      200  {object}  experiment.Experiment
// @Failure      400  {object}  ErrorResponse  "Invalid experiment ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Router       /experiments/{id} [get]
func (h *ExperimentHandler) GetExperiment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("experiment.parse_id.failed",
			zap.Error(err),
			zap.String("path", r.URL.Path),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	exp, err := h.svc.GetExperiment(r.Context(), id)
	if err != nil {
		logErr := h.lg.Warn
		if errors.Is(err, experiment.ErrExperimentNotFound) {
			logErr = h.lg.Info
		} else if !errors.Is(err, context.Canceled) {
			logErr = h.lg.Error
		}
		logErr("experiment.get.failed",
			zap.Error(err),
			zap.Int("id", id),
		)
		handleErr(w, err)
		return
	}

	response := experiment.ExperimentDTO{
		ID:             exp.ID,
		Name:           exp.Name,
		Description:    exp.Description,
		Status:         exp.Status,
		TrafficPercent: exp.TrafficPercent,
		StartDate:      exp.StartDate,
		EndDate:        exp.EndDate,
	}

	writeJSON(w, http.StatusOK, response)
}

// Create godoc
// @Summary      Create experiment
// @Tags         experiments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      experiment.CreateExperimentRequest  true  "Experiment data"
// @Success      201   {object}  experiment.Experiment
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Only scientist/admin can create experiments"
// @Failure      409   {object}  ErrorResponse  "Conflict error"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /experiments [post]
func (h *ExperimentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req experiment.CreateExperimentRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("experiment.decode.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("experiment.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.CreateExperimentCommand{
		Name:           req.Name,
		Description:    req.Description,
		TrafficPercent: req.TrafficPercent,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
	}

	exp, err := h.svc.CreateExperiment(r.Context(), &cmd)
	if err != nil {
		h.lg.Error("experiment.create.failed", zap.Error(err), zap.Any("command", cmd))
		handleErr(w, err)
		return
	}

	response := experiment.ExperimentDTO{
		ID:             exp.ID,
		Name:           exp.Name,
		Description:    exp.Description,
		Status:         exp.Status,
		TrafficPercent: exp.TrafficPercent,
		StartDate:      exp.StartDate,
		EndDate:        exp.EndDate,
	}

	writeJSON(w, http.StatusCreated, response)
}

// Update godoc
// @Summary      Update experiment
// @Tags         experiments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                                  true  "Experiment ID"
// @Param        body  body      experiment.UpdateExperimentRequest   true  "Update data"
// @Success      200   {object}  experiment.Experiment
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      404   {object}  ErrorResponse  "Experiment not found"
// @Failure      409   {object}  ErrorResponse  "Conflict (e.g., invalid status transition)"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /experiments/{id} [put]
func (h *ExperimentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("experiment.update.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req experiment.UpdateExperimentRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("experiment.decode.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("experiment.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.UpdateExperimentCommand{
		Name:           req.Name,
		Description:    req.Description,
		TrafficPercent: req.TrafficPercent,
		Status:         req.Status,
		EndDate:        req.EndDate,
	}

	exp, err := h.svc.UpdateExperiment(r.Context(), id, &cmd)
	if err != nil {
		h.lg.Error("experiment.update.failed",
			zap.Error(err),
			zap.Any("command", cmd),
		)
		handleErr(w, err)
		return
	}

	response := experiment.ExperimentDTO{
		ID:             exp.ID,
		Name:           exp.Name,
		Description:    exp.Description,
		Status:         exp.Status,
		TrafficPercent: exp.TrafficPercent,
		StartDate:      exp.StartDate,
		EndDate:        exp.EndDate,
	}

	writeJSON(w, http.StatusOK, response)
}

// Delete godoc
// @Summary      Delete experiment (admin only)
// @Tags         experiments
// @Security     BearerAuth
// @Param        id  path  int  true  "Experiment ID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse  "Cannot delete non-draft experiment"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Admin role required"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Router       /experiments/{id} [delete]
func (h *ExperimentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("experiment.delete.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.DeleteExperiment(r.Context(), id); err != nil {
		h.lg.Error("experiment.delete.failed", zap.Int("id", id), zap.Error(err))
		handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Restore godoc
// @Summary      Restore experiment (scientist+)
// @Tags         experiments
// @Security     BearerAuth
// @Param        id  path  int  true  "Experiment ID"
// @Success      200 {object}  experiment.Experiment
// @Router       /experiments/{id}/restore [post]
func (h *ExperimentHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("experiment.restore.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	exp, err := h.svc.RestoreExperiment(r.Context(), id)

	if err != nil {
		h.lg.Error("experiment.restore.failed", zap.Int("id", id), zap.Error(err))
		handleErr(w, err)
		return
	}

	response := experiment.ExperimentDTO{
		ID:             exp.ID,
		Name:           exp.Name,
		Description:    exp.Description,
		Status:         exp.Status,
		TrafficPercent: exp.TrafficPercent,
		StartDate:      exp.StartDate,
		EndDate:        exp.EndDate,
	}

	writeJSON(w, http.StatusOK, response)
}

// Start godoc
// @Summary      Start experiment (validates all variant models are PRODUCTION)
// @Tags         experiments
// @Security     BearerAuth
// @Param        id  path      int  true  "Experiment ID"
// @Success      200  {object}  experiment.Experiment
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Failure      409  {object}  ErrorResponse  "Model not in production / invalid transition"
// @Router       /experiments/{id}/start [post]
func (h *ExperimentHandler) Start(w http.ResponseWriter, r *http.Request) {
	h.lifecycle(w, r, "start", h.svc.StartExperiment)
}

// Stop godoc
// @Summary      Pause experiment
// @Tags         experiments
// @Security     BearerAuth
// @Param        id  path      int  true  "Experiment ID"
// @Success      200  {object}  experiment.Experiment
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Failure      409  {object}  ErrorResponse  "Invalid status transition"
// @Router       /experiments/{id}/stop [post]
func (h *ExperimentHandler) Stop(w http.ResponseWriter, r *http.Request) {
	h.lifecycle(w, r, "stop", h.svc.StopExperiment)
}

// Resume godoc
// @Summary      Resume a paused experiment
// @Tags         experiments
// @Security     BearerAuth
// @Param        id  path      int  true  "Experiment ID"
// @Success      200  {object}  experiment.Experiment
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Failure      409  {object}  ErrorResponse  "Invalid status transition"
// @Router       /experiments/{id}/resume [post]
func (h *ExperimentHandler) Resume(w http.ResponseWriter, r *http.Request) {
	h.lifecycle(w, r, "resume", h.svc.ResumeExperiment)
}

// Finish godoc
// @Summary      Finish experiment (moves to terminal state)
// @Tags         experiments
// @Security     BearerAuth
// @Param        id  path      int  true  "Experiment ID"
// @Success      200  {object}  experiment.Experiment
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Failure      409  {object}  ErrorResponse  "Invalid status transition"
// @Router       /experiments/{id}/finish [post]
func (h *ExperimentHandler) Finish(w http.ResponseWriter, r *http.Request) {
	h.lifecycle(w, r, "finish", h.svc.FinishExperiment)
}

func (h *ExperimentHandler) lifecycle(
	w http.ResponseWriter, r *http.Request,
	operation string,
	fn func(ctx context.Context, id int) (*experiment.Experiment, error),
) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("lifecycle.invalid_id",
			zap.Error(err),
			zap.String("op", operation),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	exp, err := fn(r.Context(), id)
	if err != nil {
		h.lg.Error("lifecycle.failed",
			zap.String("operation", operation),
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.ExperimentDTO{
		ID:             exp.ID,
		Name:           exp.Name,
		Description:    exp.Description,
		Status:         exp.Status,
		TrafficPercent: exp.TrafficPercent,
		StartDate:      exp.StartDate,
		EndDate:        exp.EndDate,
	}

	writeJSON(w, http.StatusOK, response)
}

// AddVariant godoc
// @Summary      Add variant to experiment
// @Description  Validates that the referenced model exists in model registry.
// @Description  If experiment is not DRAFT, model must be in PRODUCTION status.
// @Tags         variants
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                               true  "Experiment ID"
// @Param        body  body      experiment.CreateVariantRequest  true  "Variant data"
// @Success      201   {object}  experiment.Variant
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      404   {object}  ErrorResponse  "Experiment or model not found in registry"
// @Failure      409   {object}  ErrorResponse  "Model not in production (for non-draft experiment)"
// @Failure      422   {object}  ErrorResponse  "Weights sum mismatch or not enough variants"
// @Router       /experiments/{id}/variants [post]
func (h *ExperimentHandler) AddVariant(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("variant.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req experiment.CreateVariantRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("variant.decode.failed",
			zap.Error(err),
			zap.Int("experiment_id", id),
		)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("variant.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.CreateVariantCommand{
		ExperimentID:  id,
		Name:          req.VariantName,
		ModelID:       req.ModelID,
		TrafficWeight: req.TrafficWeight,
		IsControl:     req.IsControl,
		Description:   req.Description,
	}

	v, err := h.svc.AddVariant(r.Context(), id, &cmd)
	if err != nil {
		h.lg.Error("variant.add.failed",
			zap.Int("experiment_id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.VariantDTO{
		ID:            v.ID,
		ExperimentID:  v.ExperimentID,
		Name:          v.Name,
		ModelID:       v.ModelID,
		TrafficWeight: v.TrafficWeight,
		IsControl:     v.IsControl,
		Description:   v.Description,
	}

	writeJSON(w, http.StatusCreated, response)
}

// GetVariant godoc
// @Summary      Get variant by ID
// @Tags         variants
// @Security     BearerAuth
// @Produce      json
// @Param        id  path      int  true  "Variant ID"
// @Success      200  {object}  experiment.Variant
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Variant not found"
// @Router       /variants/{id} [get]
func (h *ExperimentHandler) GetVariant(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("variant.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	v, err := h.svc.GetVariant(r.Context(), id)
	if err != nil {
		h.lg.Error("variant.get.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.VariantDTO{
		ID:            v.ID,
		ExperimentID:  v.ExperimentID,
		Name:          v.Name,
		ModelID:       v.ModelID,
		TrafficWeight: v.TrafficWeight,
		IsControl:     v.IsControl,
		Description:   v.Description,
	}

	writeJSON(w, http.StatusOK, response)
}

// UpdateVariant godoc
// @Summary      Update variant
// @Tags         variants
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                               true  "Variant ID"
// @Param        body  body      experiment.UpdateVariantRequest   true  "Update data"
// @Success      200   {object}  experiment.Variant
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      404   {object}  ErrorResponse  "Variant not found"
// @Failure      409   {object}  ErrorResponse  "Model not in production (for active experiment)"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /variants/{id} [put]
func (h *ExperimentHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("variant.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req experiment.UpdateVariantRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("variant.decode.failed",
			zap.Error(err),
			zap.Int("variant_id", id),
		)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("variant.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.UpdateVariantCommand{
		ModelID:       req.ModelID,
		TrafficWeight: req.TrafficWeight,
		IsControl:     req.IsControl,
	}

	v, err := h.svc.UpdateVariant(r.Context(), id, &cmd)
	if err != nil {
		h.lg.Error("variant.update.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.VariantDTO{
		ID:            v.ID,
		ExperimentID:  v.ExperimentID,
		Name:          v.Name,
		ModelID:       v.ModelID,
		TrafficWeight: v.TrafficWeight,
		IsControl:     v.IsControl,
		Description:   v.Description,
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteVariant godoc
// @Summary      Delete variant
// @Tags         variants
// @Security     BearerAuth
// @Param        id  path  int  true  "Variant ID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Variant not found"
// @Failure      409  {object}  ErrorResponse  "Cannot delete variant from a non-draft experiment"
// @Router       /variants/{id} [delete]
func (h *ExperimentHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("variant.parse_id.failed",
			zap.Error(err),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteVariant(r.Context(), id); err != nil {
		h.lg.Error("variant.delete.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListVariants godoc
// @Summary      List all variants of an experiment
// @Tags         variants
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Experiment ID"
// @Success      200  {array}  experiment.Variant
// @Failure      400  {object}  ErrorResponse  "Invalid experiment ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Router       /experiments/{id}/variants [get]
func (h *ExperimentHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("variant.list.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	variants, err := h.svc.ListVariants(r.Context(), id)
	if err != nil {
		h.lg.Error("variant.list.failed",
			zap.Int("experiment_id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := make([]experiment.VariantDTO, len(variants))
	for i, v := range variants {
		response[i] = experiment.VariantDTO{
			ID:            v.ID,
			ExperimentID:  v.ExperimentID,
			Name:          v.Name,
			ModelID:       v.ModelID,
			TrafficWeight: v.TrafficWeight,
			IsControl:     v.IsControl,
			Description:   v.Description,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// ListMetrics godoc
// @Summary      List all metrics (global catalog)
// @Tags         metrics
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}  experiment.Metric
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Router       /metrics [get]
func (h *ExperimentHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.svc.ListMetrics(r.Context())
	if err != nil {
		h.lg.Error("metric.list.failed", zap.Error(err))
		handleErr(w, err)
		return
	}

	response := make([]experiment.MetricDTO, len(metrics))
	for i, m := range metrics {
		response[i] = experiment.MetricDTO{
			ID:      m.ID,
			Name:    m.Name,
			Type:    m.Type,
			Formula: m.Formula,
			Unit:    m.Unit,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// GetMetric godoc
// @Summary      Get metric by ID
// @Tags         metrics
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Metric ID"
// @Success      200  {object}  experiment.Metric
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Metric not found"
// @Router       /metrics/{id} [get]
func (h *ExperimentHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("metric.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	m, err := h.svc.GetMetric(r.Context(), id)
	if err != nil {
		h.lg.Error("metric.get.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.MetricDTO{
		ID:      m.ID,
		Name:    m.Name,
		Type:    m.Type,
		Formula: m.Formula,
		Unit:    m.Unit,
	}

	writeJSON(w, http.StatusOK, response)
}

// CreateMetric godoc
// @Summary      Create a new metric
// @Tags         metrics
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      experiment.CreateMetricRequest  true  "Metric data"
// @Success      201   {object}  experiment.Metric
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      409   {object}  ErrorResponse  "Metric already exists"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /metrics [post]
func (h *ExperimentHandler) CreateMetric(w http.ResponseWriter, r *http.Request) {
	var req experiment.CreateMetricRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("metric.decode.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("metric.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.CreateMetricCommand{
		Name:    req.Name,
		Type:    req.Type,
		Formula: req.Formula,
		Unit:    req.Unit,
	}

	m, err := h.svc.CreateMetric(r.Context(), cmd)
	if err != nil {
		h.lg.Error("metric.create.failed",
			zap.Error(err),
			zap.String("name", cmd.Name),
		)
		handleErr(w, err)
		return
	}

	response := experiment.MetricDTO{
		ID:      m.ID,
		Name:    m.Name,
		Type:    m.Type,
		Formula: m.Formula,
		Unit:    m.Unit,
	}

	writeJSON(w, http.StatusCreated, response)
}

// UpdateMetric godoc
// @Summary      Update an existing metric
// @Tags         metrics
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                               true  "Metric ID"
// @Param        body  body      experiment.UpdateMetricRequest   true  "Update data"
// @Success      200   {object}  experiment.Metric
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      404   {object}  ErrorResponse  "Metric not found"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /metrics/{id} [put]
func (h *ExperimentHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("metric.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req experiment.UpdateMetricRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("metric.decode.failed",
			zap.Error(err),
			zap.Int("id", id),
		)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("metric.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.UpdateMetricCommand{
		Name:    req.Name,
		Type:    req.Type,
		Formula: req.Formula,
		Unit:    req.Unit,
	}

	m, err := h.svc.UpdateMetric(r.Context(), id, cmd)
	if err != nil {
		h.lg.Error("metric.update.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.MetricDTO{
		ID:      m.ID,
		Name:    m.Name,
		Type:    m.Type,
		Formula: m.Formula,
		Unit:    m.Unit,
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteMetric godoc
// @Summary      Delete a metric (admin only)
// @Tags         metrics
// @Security     BearerAuth
// @Param        id  path  int  true  "Metric ID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Admin role required"
// @Failure      404  {object}  ErrorResponse  "Metric not found"
// @Failure      409  {object}  ErrorResponse  "Metric is attached to an experiment"
// @Router       /metrics/{id} [delete]
func (h *ExperimentHandler) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("metric.parse_id.failed",
			zap.Error(err),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteMetric(r.Context(), id); err != nil {
		h.lg.Error("metric.delete.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AttachMetric godoc
// @Summary      Attach a metric to an experiment
// @Tags         experiment-metrics
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                           true  "Experiment ID"
// @Param        body  body      experiment.AttachMetricRequest  true  "Metric attachment data"
// @Success      201   {object}  experiment.ExperimentMetric
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      404   {object}  ErrorResponse  "Experiment or metric not found"
// @Failure      409   {object}  ErrorResponse  "Metric already attached to experiment"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /experiments/{id}/metrics [post]
func (h *ExperimentHandler) AttachMetric(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("metric.attach.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req experiment.AttachMetricRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("metric.attach.decode.failed",
			zap.Error(err),
			zap.Int("experiment_id", id),
		)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("metric.attach.validation.failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := experiment.AttachMetricCommand{
		MetricID: req.MetricID,
		Goal:     req.Goal,
	}

	m, err := h.svc.AttachMetric(r.Context(), id, cmd)
	if err != nil {
		h.lg.Error("metric.attach.failed",
			zap.Int("experiment_id", id),
			zap.Int("metric_id", cmd.MetricID),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := experiment.ExperimentMetricDTO{
		MetricID:     m.MetricID,
		ExperimentID: m.ExperimentID,
		Goal:         m.Goal,
	}

	writeJSON(w, http.StatusCreated, response)
}

// ListExperimentMetrics godoc
// @Summary      List metrics attached to an experiment
// @Tags         experiment-metrics
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  int  true  "Experiment ID"
// @Success      200  {array}  experiment.ExperimentMetric
// @Failure      400  {object}  ErrorResponse  "Invalid experiment ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment not found"
// @Router       /experiments/{id}/metrics [get]
func (h *ExperimentHandler) ListExperimentMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("metric.list_experiment.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics, err := h.svc.ListExperimentMetrics(r.Context(), id)
	if err != nil {
		h.lg.Error("metric.list_experiment.failed",
			zap.Int("experiment_id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := make([]experiment.ExperimentMetricDTO, len(metrics))
	for i, m := range metrics {
		response[i] = experiment.ExperimentMetricDTO{
			MetricID:     m.MetricID,
			ExperimentID: m.ExperimentID,
			Goal:         m.Goal,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// DetachMetric godoc
// @Summary      Detach a metric from an experiment
// @Tags         experiment-metrics
// @Security     BearerAuth
// @Param        id        path  int  true  "Experiment ID"
// @Param        metricId  path  int  true  "Metric ID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse  "Invalid ID(s)"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Experiment or metric not found, or metric not attached"
// @Router       /experiments/{id}/metrics/{metricId} [delete]
func (h *ExperimentHandler) DetachMetric(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("metric.detach.parse_id.failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metricID, err := parseID(r, "metricId")
	if err != nil {
		h.lg.Warn("metric.detach.parse_metric_id.failed",
			zap.Error(err),
			zap.Int("experiment_id", id),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DetachMetric(r.Context(), id, metricID); err != nil {
		h.lg.Error("metric.detach.failed",
			zap.Int("experiment_id", id),
			zap.Int("metric_id", metricID),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
