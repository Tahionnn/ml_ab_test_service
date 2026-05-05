package domain

import (
	"context"
)

type Splitter interface {
	GetVariant(ctx context.Context, req VariantRequest) (VariantResponse, error)
}
