package allocation

import "github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain"

type Allocator interface {
	AllocateVariants(bucket int, variants []domain.VariantConfig) (domain.VariantConfig, error)
}
