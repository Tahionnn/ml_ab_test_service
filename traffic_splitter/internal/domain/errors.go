package domain

import "errors"

var (
	ErrNoActiveExperiments = errors.New("no active experiments")
	ErrConfigNotFound      = errors.New("experiment config not found")
	ErrNoControlVariant    = errors.New("no control variant")

	ErrNoVariants        = errors.New("no variants configured")
	ErrNoMatchingVariant = errors.New("no matching variant")

	ErrUpstreamUnavailable = errors.New("upstream service unavailable")
)
