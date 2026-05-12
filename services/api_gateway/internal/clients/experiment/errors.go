package experiment

import (
	"errors"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/apierror"
)

var (
	// 404
	ErrExperimentNotFound = errors.New("experiment not found")
	ErrMetricNotFound     = errors.New("metric not found")
	ErrVariantNotFound    = errors.New("variant not found")
	ErrMetricNotAttached  = errors.New("metric not attached")

	// 400
	ErrInvalidStatusTransition   = errors.New("invalid status transition")
	ErrExperimentCannotBeDeleted = errors.New("experiment cannot be deleted")
	ErrMetricAlreadyAttached     = errors.New("metric already attached")

	//409
	ErrExperimentAlreadyExists = errors.New("experimentt already exists")

	// 422
	ErrWeightsDoNotSumTo100 = errors.New("weights do not sum to 100")
	ErrNotEnoughVariants    = errors.New("not enough variants")
	ErrValidationError      = errors.New("validation error")
)

func mapResponseError(apiErr *apierror.APIError) error {
	return apiErr
}
