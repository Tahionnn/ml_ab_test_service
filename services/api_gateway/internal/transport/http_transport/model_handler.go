package httptransport

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	modelsservice "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/models_service"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
	mw "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport/middleware"
	"go.uber.org/zap"
)

// ─── MODEL HANDLER ──────────────────────────────────────────────────────────

type ModelHandler struct {
	svc *gateway.Service
	lg  *zap.Logger
}

func NewModelHandler(svc *gateway.Service, lg *zap.Logger) *ModelHandler {
	return &ModelHandler{svc: svc, lg: lg}
}

func (h *ModelHandler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	scientist := mw.RequireRole(user.Scientist, user.Admin)
	admin := mw.RequireRole(user.Admin)

	r.Group(func(r chi.Router) {
		r.Use(auth)

		r.Get("/models", h.List)
		r.Get("/models/{id}", h.Get)
		r.With(scientist).Post("/models", h.Create)
		r.With(scientist).Put("/models/{id}", h.Update)
		r.With(admin).Delete("/models/{id}", h.Delete)
		r.With(scientist).Patch("/models/{id}/status", h.ChangeStatus)
		r.With(scientist).Post("/models/{id}/deploy", h.Deploy)
		r.With(scientist).Post("/models/{id}/undeploy", h.Undeploy)
	})
}

// List godoc
// @Summary      List all models
// @Tags         models
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   modelsservice.Model
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Router       /models [get]
func (h *ModelHandler) List(w http.ResponseWriter, r *http.Request) {
	models, err := h.svc.ListModels(r.Context())
	if err != nil {
		h.lg.Error("model.list.failed", zap.Error(err))
		handleErr(w, err)
		return
	}

	response := make([]modelsservice.ModelResponse, len(models))
	for i, m := range models {
		response[i] = modelsservice.ModelResponse{
			ID:              *m.ID,
			Name:            m.Name,
			Version:         m.Version,
			ArtifactURI:     m.ArtifactUrl,
			ServingEndpoint: m.ServingEndpoint,
			Framework:       m.Framework,
			Status:          m.Status,
			DeploymentID:    m.DeploymentID,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// Get godoc
// @Summary      Get model by ID
// @Tags         models
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "Model ID"
// @Success      200  {object}  modelsservice.Model
// @Failure      400  {object}  ErrorResponse  "Invalid model ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "Model not found"
// @Router       /models/{id} [get]
func (h *ModelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("model.get.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	m, err := h.svc.GetModel(r.Context(), id)
	if err != nil {
		h.lg.Error("model.get.failed", zap.Int("id", id), zap.Error(err))
		handleErr(w, err)
		return
	}

	response := modelsservice.ModelResponse{
		ID:              *m.ID,
		Name:            m.Name,
		Version:         m.Version,
		ArtifactURI:     m.ArtifactUrl,
		ServingEndpoint: m.ServingEndpoint,
		Framework:       m.Framework,
		Status:          m.Status,
		DeploymentID:    m.DeploymentID,
	}

	writeJSON(w, http.StatusOK, response)
}

// Create godoc
// @Summary      Register a new model
// @Tags         models
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      modelsservice.CreateModelRequest  true  "Model data"
// @Success      201   {object}  modelsservice.Model
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      409   {object}  ErrorResponse  "Model with same name and version already exists"
// @Failure      422   {object}  ErrorResponse  "Validation error (invalid name, version, artifact URI, or framework)"
// @Router       /models [post]
func (h *ModelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req modelsservice.CreateModelRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("model.create.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("model.create.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := modelsservice.CreateModelCommand{
		Name:        req.Name,
		Version:     req.Version,
		ArtifactURI: req.ArtifactURI,
		Framework:   req.Framework,
		Status:      req.Status,
	}

	m, err := h.svc.CreateModel(r.Context(), cmd)
	if err != nil {
		h.lg.Error("model.create.failed",
			zap.Error(err),
			zap.String("name", cmd.Name),
		)
		handleErr(w, err)
		return
	}

	response := modelsservice.ModelResponse{
		ID:              *m.ID,
		Name:            m.Name,
		Version:         m.Version,
		ArtifactURI:     m.ArtifactUrl,
		ServingEndpoint: m.ServingEndpoint,
		Framework:       m.Framework,
		Status:          m.Status,
		DeploymentID:    m.DeploymentID,
	}

	writeJSON(w, http.StatusCreated, response)
}

// Update godoc
// @Summary      Update an existing model (scientist+)
// @Tags         models
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                               true  "Model ID"
// @Param        body  body      modelsservice.UpdateModelRequest  true  "Model update data (partial)"
// @Success      200   {object}  modelsservice.Model
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden (scientist or admin required)"
// @Failure      404   {object}  ErrorResponse  "Model not found"
// @Failure      409   {object}  ErrorResponse  "Model cannot be updated (e.g., production model locked)"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /models/{id} [put]
func (h *ModelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("model.update.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req modelsservice.UpdateModelRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("model.update.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("model.update.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := modelsservice.UpdateModelCommand{
		Name:        req.Name,
		Version:     req.Version,
		ArtifactURI: req.ArtifactURI,
		Framework:   req.Framework,
		Status:      req.Status,
	}

	m, err := h.svc.UpdateModel(r.Context(), id, cmd)
	if err != nil {
		h.lg.Error("model.update.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := modelsservice.ModelResponse{
		ID:              *m.ID,
		Name:            m.Name,
		Version:         m.Version,
		ArtifactURI:     m.ArtifactUrl,
		ServingEndpoint: m.ServingEndpoint,
		Framework:       m.Framework,
		Status:          m.Status,
		DeploymentID:    m.DeploymentID,
	}

	writeJSON(w, http.StatusOK, response)
}

// Delete godoc
// @Summary      Delete a model (admin only)
// @Tags         models
// @Security     BearerAuth
// @Param        id  path  int  true  "Model ID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Admin role required"
// @Failure      404  {object}  ErrorResponse  "Model not found"
// @Failure      409  {object}  ErrorResponse  "Model cannot be deleted (used in active experiment)"
// @Router       /models/{id} [delete]
func (h *ModelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("model.delete.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteModel(r.Context(), id); err != nil {
		h.lg.Error("model.delete.failed", zap.Int("id", id), zap.Error(err))
		handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangeStatus godoc
// @Summary      Change model status manually (scientist+)
// @Description  Allowed status transitions depend on current state and deployment status.
// @Tags         models
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                                   true  "Model ID"
// @Param        body  body      modelsservice.ChangeStatusRequest    true  "New status (staging, archived)"
// @Success      200   {object}  modelsservice.Model
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Forbidden"
// @Failure      404   {object}  ErrorResponse  "Model not found"
// @Failure      409   {object}  ErrorResponse  "Invalid status transition"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /models/{id}/status [patch]
func (h *ModelHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("model.change_status.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req modelsservice.ChangeStatusRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("model.change_status.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("model.change_status.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := modelsservice.ChangeStatusCommand{
		ModelID: id,
		Status:  req.Status,
	}

	m, err := h.svc.ChangeModelStatus(r.Context(), id, cmd)
	if err != nil {
		h.lg.Error("model.change_status.failed",
			zap.Int("id", id),
			zap.Error(err),
		)
		handleErr(w, err)
		return
	}

	response := modelsservice.ModelResponse{
		ID:              *m.ID,
		Name:            m.Name,
		Version:         m.Version,
		ArtifactURI:     m.ArtifactUrl,
		ServingEndpoint: m.ServingEndpoint,
		Framework:       m.Framework,
		Status:          m.Status,
		DeploymentID:    m.DeploymentID,
	}

	writeJSON(w, http.StatusOK, response)
}

// Deploy godoc
// @Summary      Deploy model to production (async via Kafka)
// @Description  Initiates asynchronous deployment. Model status will change to 'pending', eventually 'production' or 'failed'.
// @Tags         models
// @Security     BearerAuth
// @Param        id  path      int  true  "Model ID"
// @Success      202 {object}  modelsservice.DeploymentResult  "Deployment accepted (async)"
// @Failure      400 {object}  ErrorResponse  "Invalid model ID"
// @Failure      401 {object}  ErrorResponse  "Unauthorized"
// @Failure      403 {object}  ErrorResponse  "Forbidden (scientist+ required)"
// @Failure      404 {object}  ErrorResponse  "Model not found"
// @Failure      409 {object}  ErrorResponse  "Model cannot be deployed (e.g., already deployed or invalid state)"
// @Router       /models/{id}/deploy [post]
func (h *ModelHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("model.deploy.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.svc.DeployModel(r.Context(), id)
	if err != nil {
		h.lg.Error("model.deploy.failed", zap.Int("id", id), zap.Error(err))
		handleErr(w, err)
		return
	}

	response := modelsservice.DeployResponse{
		DeploymentID: res.ID,
		Status:       string(res.Status),
	}

	writeJSON(w, http.StatusAccepted, response)
}

// Undeploy godoc
// @Summary      Undeploy model from production
// @Description  Removes model from serving and archives it.
// @Tags         models
// @Security     BearerAuth
// @Param        id  path      int  true  "Model ID"
// @Success      202 {object}  modelsservice.DeploymentResult  "Undeployment accepted (async)"
// @Failure      400 {object}  ErrorResponse  "Invalid model ID"
// @Failure      401 {object}  ErrorResponse  "Unauthorized"
// @Failure      403 {object}  ErrorResponse  "Forbidden (scientist+ required)"
// @Failure      404 {object}  ErrorResponse  "Model not found"
// @Failure      409 {object}  ErrorResponse  "Model cannot be undeployed (e.g., not deployed or used in active experiment)"
// @Router       /models/{id}/undeploy [post]
func (h *ModelHandler) Undeploy(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("model.undeploy.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.svc.UndeployModel(r.Context(), id)
	if err != nil {
		h.lg.Error("model.undeploy.failed", zap.Int("id", id), zap.Error(err))
		handleErr(w, err)
		return
	}

	response := modelsservice.UndeployResponse{
		Success: true,
		Message: fmt.Sprintf("Deploy: ID=%s, Status=%v", res.ID, res.Status),
	}

	writeJSON(w, http.StatusAccepted, response)
}
