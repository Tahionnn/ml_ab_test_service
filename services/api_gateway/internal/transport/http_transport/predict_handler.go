package httptransport

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
	mw "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport/middleware"
	"go.uber.org/zap"
)

type PredictHandler struct {
	svc *gateway.Service
	lg  *zap.Logger
}

func NewPredictHandler(svc *gateway.Service, lg *zap.Logger) *PredictHandler {
	return &PredictHandler{svc: svc, lg: lg}
}

func (h *PredictHandler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(auth)
		r.Post("/predict", h.Predict)
		r.Post("/simulate", h.Simulate)
	})
}

// Predict godoc
// @Summary      Get ML prediction with A/B routing
// @Description  Routes request through traffic splitter to the appropriate model variant.
// @Description  Publishes a prediction event to Kafka asynchronously.
// @Tags         predict
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      gateway.PredictRequest   true  "Predict request"
// @Success      200   {object}  gateway.PredictResponse
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      404   {object}  ErrorResponse  "Experiment or variant not found"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Failure      503   {object}  ErrorResponse  "Splitter or serving endpoint unavailable"
// @Router       /predict [post]
func (h *PredictHandler) Predict(w http.ResponseWriter, r *http.Request) {
	var req gateway.PredictRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("predict.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("predict.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.UserID == "" {
		if id, ok := mw.GetUserID(r); ok {
			req.UserID = fmt.Sprint(id)
		}
	}

	resp, err := h.svc.Predict(r.Context(), req)
	if err != nil {
		h.lg.Error("predict.failed", zap.Error(err))
		handleErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Simulate godoc
// @Summary      Simulate user requests for A/B testing (no real predictions)
// @Description  Sends multiple users through the splitter and publishes events to Kafka.
// @Description  Use this to generate test data when you don't have a real frontend.
// @Tags         predict
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      gateway.SimulateRequest   true  "Simulation params"
// @Success      200   {array}   gateway.SimulateResult
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      404   {object}  ErrorResponse  "Experiment not found"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /simulate [post]
func (h *PredictHandler) Simulate(w http.ResponseWriter, r *http.Request) {
	var req gateway.SimulateRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("simulate.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("simulate.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	results, err := h.svc.Simulate(r.Context(), req)
	if err != nil {
		h.lg.Error("simulate.failed", zap.Error(err))
		handleErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, results)
}
