package updater

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain"
	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain/cache"
	experimentclient "github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain/experiment_client"
	modelclient "github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain/model_client"
	"go.uber.org/zap"
)

type CacheLoader struct {
	expClient   experimentclient.ExperimentClient
	modelClient modelclient.ModelClient
	store       cache.ConfigStore
	logger      *zap.Logger
}

func NewCacheLoader(exp experimentclient.ExperimentClient, model modelclient.ModelClient, store cache.ConfigStore, logger *zap.Logger) *CacheLoader {
	return &CacheLoader{expClient: exp, modelClient: model, store: store, logger: logger}
}

func (l *CacheLoader) LoadAll(ctx context.Context) error {
	experiments, err := l.expClient.GetActiveExperiments(ctx)
	if err != nil {
		return err
	}

	if len(experiments) == 0 {
		l.logger.Info("no active experiments, clearing cache")
		l.store.SetAll(map[string]domain.ExperimentConfig{})
		return domain.ErrNoActiveExperiments
	}

	modelIDs := collectModelIDs(experiments)

	if len(modelIDs) == 0 {
		l.logger.Info("no model ids found, clearing cache")
		l.store.SetAll(map[string]domain.ExperimentConfig{})
		return domain.ErrNoActiveExperiments
	}

	endpointsMap, err := l.modelClient.GetEndpointsBulk(ctx, modelIDs)
	if err != nil {
		return fmt.Errorf("failed to get bulk endpoints: %w", err)
	}

	configs := make(map[string]domain.ExperimentConfig, len(experiments))

	for _, exp := range experiments {
		config := toExperimentConfig(exp, endpointsMap)
		configs[config.ID] = config
	}

	l.store.SetAll(configs)
	return nil
}

func toExperimentConfig(dto experimentclient.ExperimentDTO, endpoints map[string]string) domain.ExperimentConfig {
	config := domain.ExperimentConfig{
		ID:         strconv.Itoa(dto.ID),
		Name:       dto.Name,
		Status:     dto.Status,
		TrafficPct: dto.TrafficPercent,
		Variants:   make([]domain.VariantConfig, 0, len(dto.Variants)),
	}

	const scale = domain.TotalBuckets / 100

	currentBucket := 0
	for _, v := range dto.Variants {
		modelIDStr := strconv.Itoa(v.ModelID)
		bucketEnd := currentBucket + v.TrafficWeight*scale

		varCfg := domain.VariantConfig{
			ID:              strconv.Itoa(v.ID),
			Name:            v.Name,
			ModelId:         modelIDStr,
			ServingEndpoint: endpoints[modelIDStr],
			Weight:          v.TrafficWeight,
			IsControl:       v.IsControl,
			BucketStart:     currentBucket,
			BucketEnd:       bucketEnd,
		}

		config.Variants = append(config.Variants, varCfg)
		currentBucket = bucketEnd
	}

	return config
}

func collectModelIDs(experiments []experimentclient.ExperimentDTO) []string {
	uniqueIDs := make(map[string]struct{})
	for _, exp := range experiments {
		for _, v := range exp.Variants {
			uniqueIDs[strconv.Itoa(v.ModelID)] = struct{}{}
		}
	}

	ids := make([]string, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		ids = append(ids, id)
	}
	return ids
}
