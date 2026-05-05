package grpctransport

import (
	"context"

	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
	pb "github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/pb/splitter/v1"
)

type TrafficSplitter struct {
	pb.UnimplementedTrafficSplitterServer
	splitter domain.Splitter
}

func NewHandler(splitter domain.Splitter) *TrafficSplitter {
	return &TrafficSplitter{
		splitter: splitter,
	}
}

func (s *TrafficSplitter) GetVariant(ctx context.Context, req *pb.GetVariantRequest) (*pb.GetVariantResponse, error) {
	res, err := s.splitter.GetVariant(ctx, domain.VariantRequest{
		UserId:       req.UserId,
		ExperimentID: req.ExperimentId,
		Attributes:   req.Attributes,
	})

	if err != nil {
		return nil, err
	}

	return &pb.GetVariantResponse{
		VariantId:        res.VariantId,
		ServingEndpoint:  res.ServingEndpoint,
		IsControl:        res.IsControl,
		AssignmentReason: res.AssignmentReason,
	}, nil
}

func (s *TrafficSplitter) GetVariants(ctx context.Context, req *pb.GetVariantsRequest) (*pb.GetVariantsResponse, error) {
	result := make(map[string]*pb.GetVariantResponse)

	for _, expID := range req.GetExperimentIds() {
		resp, err := s.splitter.GetVariant(ctx, domain.VariantRequest{
			UserId:       req.GetUserId(),
			ExperimentID: expID,
		})

		if err != nil {
			return nil, err
		}

		result[expID] = &pb.GetVariantResponse{
			VariantId:        resp.VariantId,
			ServingEndpoint:  resp.ServingEndpoint,
			IsControl:        resp.IsControl,
			AssignmentReason: resp.AssignmentReason,
		}
	}

	return &pb.GetVariantsResponse{
		Assignments: result,
	}, nil
}
