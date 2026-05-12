package apierror

import (
	"errors"
	"fmt"
)

type APIError struct {
	StatusCode int
	Message    string
	Path       string
	Service    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] API error: %d - %s (path: %s)", e.Service, e.StatusCode, e.Message, e.Path)
}

func IsStatus(err error, code int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == code
	}
	return false
}
