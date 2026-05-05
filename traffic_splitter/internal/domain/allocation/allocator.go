package allocation

import "github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"

type Allocator interface {
	AllocateVariants(bucket int, variants []domain.VariantConfig) (domain.VariantConfig, error)
}
