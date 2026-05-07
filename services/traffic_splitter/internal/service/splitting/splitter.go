package splitting

import (
	"context"
	"errors"

	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain"
	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain/allocation"
	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain/cache"
	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/hasher"
)

type TrafficSplitter struct {
	store     cache.ConfigStore
	hasher    hasher.Hasher
	allocator allocation.Allocator
}

func NewSplitter(store cache.ConfigStore, hasher hasher.Hasher, allocator allocation.Allocator) *TrafficSplitter {
	return &TrafficSplitter{
		store:     store,
		hasher:    hasher,
		allocator: allocator,
	}
}

func (ts *TrafficSplitter) GetVariant(ctx context.Context, req domain.VariantRequest) (domain.VariantResponse, error) {
	config, ok := ts.store.Get(ctx, req.ExperimentID)
	if !ok {
		return ts.getFallbackVariant("config_not_found"), nil
	}

	if config.Status != domain.StatusRunning {
		return ts.getControlVariant(ctx, config)
	}

	if config.TrafficPct < 100 {
		expBucket := ts.getBucket(req.UserId, "gate:"+req.ExperimentID)
		gateThreshold := config.TrafficPct * 100
		if expBucket >= gateThreshold {
			return ts.getFallbackVariant("traffic_gate"), nil
		}
	}

	bucket := ts.getBucket(req.UserId, req.ExperimentID)

	variant, err := ts.allocator.AllocateVariants(bucket, config.Variants)
	if err != nil {
		if errors.Is(err, domain.ErrNoVariants) || errors.Is(err, domain.ErrNoMatchingVariant) {
			return ts.getFallbackVariant("allocation_failed"), nil
		}
		return ts.getFallbackVariant("allocation_failed"), err
	}

	return ts.resolveVariant(ctx, variant)
}

func (ts *TrafficSplitter) getControlVariant(ctx context.Context, config domain.ExperimentConfig) (domain.VariantResponse, error) {
	for _, v := range config.Variants {
		if v.IsControl {
			return domain.VariantResponse{
				VariantId:        v.ID,
				ServingEndpoint:  v.ServingEndpoint,
				IsControl:        true,
				AssignmentReason: "control_variant",
			}, nil
		}
	}
	return ts.getFallbackVariant("no_control_variant"), nil
}

func (ts *TrafficSplitter) getFallbackVariant(reason string) domain.VariantResponse {
	return domain.VariantResponse{
		VariantId:        "default",
		ServingEndpoint:  "",
		IsControl:        true,
		AssignmentReason: reason,
	}
}

func (ts *TrafficSplitter) getBucket(userID, experimentID string) int {
	key := experimentID + ":" + userID
	h := ts.hasher.Hash([]byte(key))
	return int(h % domain.TotalBuckets)
}

func (ts *TrafficSplitter) resolveVariant(
	ctx context.Context,
	variant domain.VariantConfig,
) (domain.VariantResponse, error) {
	return domain.VariantResponse{
		VariantId:        variant.ID,
		ServingEndpoint:  variant.ServingEndpoint,
		IsControl:        variant.IsControl,
		AssignmentReason: "allocated",
	}, nil
}
