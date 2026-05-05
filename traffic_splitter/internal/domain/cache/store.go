package cache

import (
	"context"

	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
)

type ConfigStore interface {
	Get(ctx context.Context, experimentID string) (domain.ExperimentConfig, bool)
	Set(experimentID string, config domain.ExperimentConfig)
	SetAll(configs map[string]domain.ExperimentConfig)
	UpdateModelEndpoint(modelID string, endpoint string)
}
