package splitter

import (
	"context"
	"fmt"
	"time"

	splitterv1 "github.com/tahion/ml_ab_test_service/gen/go/splitter/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SplitterService interface {
	GetVariant(ctx context.Context, userID, experimentID string, attrs map[string]string) (*Variant, error)
	GetVariants(ctx context.Context, userID string, experimentIDs []string) (map[string]*Variant, error)
}

type GRPCSplitterClient struct {
	client splitterv1.TrafficSplitterClient
	conn   *grpc.ClientConn
	config SplitterConfig
}

type SplitterConfig struct {
	Address string
	Timeout time.Duration
}

func NewGRPCClient(cfg SplitterConfig) (*GRPCSplitterClient, error) {
	conn, err := grpc.NewClient(cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &GRPCSplitterClient{
		client: splitterv1.NewTrafficSplitterClient(conn),
		conn:   conn,
		config: cfg,
	}, nil
}

func mapVariant(resp *splitterv1.GetVariantResponse) *Variant {
	if resp == nil {
		return nil
	}

	return &Variant{
		VariantID:        resp.VariantId,
		ServingEndpoint:  resp.ServingEndpoint,
		IsControl:        resp.IsControl,
		AssignmentReason: resp.AssignmentReason,
	}
}

var _ SplitterService = (*GRPCSplitterClient)(nil)

func (c *GRPCSplitterClient) GetVariant(ctx context.Context, userID, experimentID string, attrs map[string]string) (*Variant, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	resp, err := c.client.GetVariant(ctx, &splitterv1.GetVariantRequest{
		UserId:       userID,
		ExperimentId: experimentID,
		Attributes:   attrs,
	})
	if err != nil {
		return nil, mapGRPCError(err)
	}

	return mapVariant(resp), nil

}
func (c *GRPCSplitterClient) GetVariants(ctx context.Context, userID string, experimentIDs []string) (map[string]*Variant, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	resp, err := c.client.GetVariants(ctx, &splitterv1.GetVariantsRequest{
		UserId:        userID,
		ExperimentIds: experimentIDs,
	})
	if err != nil {
		return nil, mapGRPCError(err)
	}

	result := make(map[string]*Variant, len(resp.Assignments))

	for expID, variantResp := range resp.Assignments {
		result[expID] = mapVariant(variantResp)
	}

	return result, nil
}

func (c *GRPCSplitterClient) Close() error {
	return c.conn.Close()
}
