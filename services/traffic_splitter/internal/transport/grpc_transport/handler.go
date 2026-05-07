package grpctransport

import (
	"context"

	splitterv1 "github.com/tahion/ml_ab_test_service/gen/go/splitter/v1"
	"github.com/tahion/ml_ab_test_service/services/traffic_splitter/internal/domain"
)

type TrafficSplitter struct {
	splitterv1.UnimplementedTrafficSplitterServer
	splitter domain.Splitter
}

func NewHandler(splitter domain.Splitter) *TrafficSplitter {
	return &TrafficSplitter{
		splitter: splitter,
	}
}

func (s *TrafficSplitter) GetVariant(ctx context.Context, req *splitterv1.GetVariantRequest) (*splitterv1.GetVariantResponse, error) {
	res, err := s.splitter.GetVariant(ctx, domain.VariantRequest{
		UserId:       req.UserId,
		ExperimentID: req.ExperimentId,
		Attributes:   req.Attributes,
	})

	if err != nil {
		return nil, err
	}

	return &splitterv1.GetVariantResponse{
		VariantId:        res.VariantId,
		ServingEndpoint:  res.ServingEndpoint,
		IsControl:        res.IsControl,
		AssignmentReason: res.AssignmentReason,
	}, nil
}

func (s *TrafficSplitter) GetVariants(ctx context.Context, req *splitterv1.GetVariantsRequest) (*splitterv1.GetVariantsResponse, error) {
	result := make(map[string]*splitterv1.GetVariantResponse)

	for _, expID := range req.GetExperimentIds() {
		resp, err := s.splitter.GetVariant(ctx, domain.VariantRequest{
			UserId:       req.GetUserId(),
			ExperimentID: expID,
		})

		if err != nil {
			return nil, err
		}

		result[expID] = &splitterv1.GetVariantResponse{
			VariantId:        resp.VariantId,
			ServingEndpoint:  resp.ServingEndpoint,
			IsControl:        resp.IsControl,
			AssignmentReason: resp.AssignmentReason,
		}
	}

	return &splitterv1.GetVariantsResponse{
		Assignments: result,
	}, nil
}
