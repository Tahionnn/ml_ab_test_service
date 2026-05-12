package gateway

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden: insufficient role")
	ErrInvalidCreds = errors.New("invalid credentials")
	ErrInvalidToken = errors.New("invalid or expired token")
)

var ErrValidation = errors.New("validation error")

var (
	ErrModelNotFound      = errors.New("model not found")
	ErrModelNotProduction = errors.New("model must be in production status to be used in a running experiment")
)

var (
	ErrSplitterUnavailable = errors.New("traffic splitter unavailable")
	ErrServingUnavailable  = errors.New("model serving endpoint unavailable")
	ErrNoActiveExperiment  = errors.New("no active experiment found for this id")
)
