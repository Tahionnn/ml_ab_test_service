package user

import (
	"errors"
	"fmt"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/apierror"
)

var (
	// 404
	ErrUserNotFound error = errors.New("user not found")

	// 400
	ErrInvalidToken       error = errors.New("invalid token")
	ErrInvalidCredentials error = errors.New("invalid credentials")

	// 403
	ErrForbidden error = errors.New("forbidden")

	// 409
	ErrUserAlreadyExists error = errors.New("user already exists")

	// 422
	ErrValidationError error = errors.New("validation error")
)

type ForbiddenError struct {
	Reason string
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Reason)
}

func (e *ForbiddenError) Unwrap() error {
	return ErrForbidden
}

func mapResponseError(apiErr *apierror.APIError) error {
	return apiErr
}
