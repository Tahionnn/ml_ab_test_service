package modelsservice

import (
	"errors"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/apierror"
)

var (
	// 404
	ErrModelNotFound = errors.New("model not found")

	// 400
	ErrInvalidEndpointHandler = errors.New("invalid endpoint handler")

	// 409
	ErrModelCannotBeUpdated    = errors.New("model cannot be updated")
	ErrModelCannotBeDeleted    = errors.New("model cannot be deleted")
	ErrModelCannotBeDeploy     = errors.New("model cannot be deployed")
	ErrModelCannotBeUndeployed = errors.New("model cannot be undeployed")

	// 422
	ErrValidationError = errors.New("validation error")
)

func mapResponseError(apiErr *apierror.APIError) error {
	return apiErr
}
