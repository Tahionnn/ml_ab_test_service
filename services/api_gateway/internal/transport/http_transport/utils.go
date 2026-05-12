package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/apierror"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, dst any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}
	return nil
}

func parseID(r *http.Request, key string) (int, error) {
	id, err := strconv.Atoi(chi.URLParam(r, key))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id: must be a positive integer")
	}
	return id, nil
}

func handleErr(w http.ResponseWriter, err error) {
	var apiErr *apierror.APIError
	if errors.As(err, &apiErr) {
		writeError(w, apiErr.StatusCode, apiErr.Message)
		return
	}

	switch {
	case errors.Is(err, gateway.ErrUnauthorized),
		errors.Is(err, gateway.ErrInvalidToken),
		errors.Is(err, gateway.ErrInvalidCreds):
		writeError(w, http.StatusUnauthorized, err.Error())

	case errors.Is(err, gateway.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())

	case errors.Is(err, gateway.ErrValidation):
		writeError(w, http.StatusUnprocessableEntity, err.Error())

	case errors.Is(err, gateway.ErrModelNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, gateway.ErrModelNotProduction):
		writeError(w, http.StatusConflict, err.Error())

	case errors.Is(err, gateway.ErrSplitterUnavailable),
		errors.Is(err, gateway.ErrServingUnavailable):
		writeError(w, http.StatusServiceUnavailable, err.Error())

	case errors.Is(err, gateway.ErrNoActiveExperiment):
		writeError(w, http.StatusNotFound, err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
