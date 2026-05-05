package allocation

import (
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
)

type allocator struct{}

func NewAllocator() Allocator {
	return &allocator{}
}

func (a *allocator) AllocateVariants(bucket int, variants []domain.VariantConfig) (domain.VariantConfig, error) {
	if len(variants) == 0 {
		return domain.VariantConfig{}, domain.ErrNoVariants
	}

	for _, v := range variants {
		if bucket >= v.BucketStart && bucket < v.BucketEnd {
			return v, nil
		}
	}

	return domain.VariantConfig{}, domain.ErrNoMatchingVariant
}
