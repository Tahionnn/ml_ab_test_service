package modelclient

import "context"

type ModelClient interface {
	GetEndpointsBulk(ctx context.Context, modelIDs []string) (map[string]string, error)
}
