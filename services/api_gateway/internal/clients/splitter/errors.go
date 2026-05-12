package splitter

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrExperimentNotFound  error = errors.New("experiment not found")
	ErrInvalidArgument     error = errors.New("invalid argument")
	ErrSplitterTimeout     error = errors.New("splitter timeout")
	ErrSplitterUnavailable error = errors.New("splitter unavailable")
)

func mapGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return ErrExperimentNotFound

	case codes.InvalidArgument:
		return ErrInvalidArgument

	case codes.DeadlineExceeded:
		return ErrSplitterTimeout

	case codes.Unavailable:
		return ErrSplitterUnavailable

	default:
		return fmt.Errorf("splitter grpc error [%s]: %s",
			st.Code(),
			st.Message(),
		)
	}
}
