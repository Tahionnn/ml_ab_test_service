package modelsservice

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/httpclient"

	"github.com/google/uuid"
)

type ModelClient interface {
	// Registry
	CreateModel(ctx context.Context, cmd CreateModelCommand) (*Model, error)
	ListModels(ctx context.Context) ([]Model, error)
	GetModel(ctx context.Context, id int) (*Model, error)
	SearchModel(ctx context.Context, query SearchModelsQuery) (*SearchResult, error)
	GetByNameVersion(ctx context.Context, name, version string) (*Model, error)
	GetByStatus(ctx context.Context, status ModelStatus) ([]Model, error)
	UpdateModel(ctx context.Context, id int, cmd UpdateModelCommand) (*Model, error)
	ChangeStatus(ctx context.Context, id int, cmd ChangeStatusCommand) (*Model, error)
	DeleteModel(ctx context.Context, id int) error

	// Serving
	DeployModel(ctx context.Context, cmd DeployCommand) (*DeploymentResult, error)
	UndeployModel(ctx context.Context, cmd UnDeployCommand) (*DeploymentResult, error)
}

type HTTPModelClient struct {
	client *httpclient.Client
}

var _ ModelClient = (*HTTPModelClient)(nil)

type Config struct {
	BaseURL   string
	Timeout   time.Duration
	AuthToken string
}

func NewHTTPModelClient(cfg Config) *HTTPModelClient {
	return &HTTPModelClient{
		client: httpclient.New(cfg.BaseURL, cfg.Timeout, cfg.AuthToken),
	}
}

func (c *HTTPModelClient) toDomainModel(dto *ModelResponse) *Model {
	var id *int
	if dto.ID != 0 {
		id = &dto.ID
	}

	return &Model{
		ID:              id,
		Name:            dto.Name,
		Version:         dto.Version,
		ArtifactUrl:     dto.ArtifactURI,
		Framework:       ModelFramework(dto.Framework),
		ServingEndpoint: dto.ServingEndpoint,
		Status:          ModelStatus(dto.Status),
		DeploymentID:    dto.DeploymentID,
	}
}

func (c *HTTPModelClient) toDomainSearchResult(dto *SearchResponse) *SearchResult {
	items := make([]Model, 0, len(dto.Models))

	for _, m := range dto.Models {
		items = append(items, *c.toDomainModel(&m))
	}

	return &SearchResult{
		Total: dto.Total,
		Items: items,
	}
}

func (c *HTTPModelClient) toDomainDeployment(dto *DeployResponse) *DeploymentResult {
	return &DeploymentResult{
		ID:     dto.DeploymentID,
		Status: ModelStatus(dto.Status),
	}
}

func (c *HTTPModelClient) toDomainUndeploy(dto *UndeployResponse) *DeploymentResult {
	status := ModelArchived
	if dto.Success {
		status = ModelArchived
	}

	return &DeploymentResult{
		ID:     uuid.Nil,
		Status: status,
	}
}

func (c *HTTPModelClient) CreateModel(ctx context.Context, cmd CreateModelCommand) (*Model, error) {
	req := CreateModelRequest{
		Name:        cmd.Name,
		Version:     cmd.Version,
		ArtifactURI: cmd.ArtifactURI,
		Framework:   cmd.Framework,
		Status:      cmd.Status,
	}

	var dto ModelResponse
	err := c.client.Do(ctx, http.MethodPost, "/model", req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainModel(&dto), nil
}

func (c *HTTPModelClient) ListModels(ctx context.Context) ([]Model, error) {
	var dtos []ModelResponse
	err := c.client.Do(ctx, http.MethodGet, "/model", nil, &dtos)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	res := make([]Model, 0, len(dtos))
	for _, v := range dtos {
		res = append(res, *c.toDomainModel(&v))
	}
	return res, nil
}

func (c *HTTPModelClient) GetModel(ctx context.Context, id int) (*Model, error) {
	var dto ModelResponse
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/model/%d", id), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainModel(&dto), nil
}

func (c *HTTPModelClient) SearchModel(ctx context.Context, query SearchModelsQuery) (*SearchResult, error) {
	req := SearchRequest{
		Q:     query.Query,
		Page:  query.Page,
		Limit: query.Limit,
	}

	var dto SearchResponse
	err := c.client.Do(ctx, http.MethodGet, "/model/search", req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainSearchResult(&dto), nil
}

func (c *HTTPModelClient) GetByNameVersion(ctx context.Context, name, version string) (*Model, error) {
	var dto ModelResponse
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/model/name/%s/version/%s", name, version), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainModel(&dto), nil
}

func (c *HTTPModelClient) GetByStatus(ctx context.Context, status ModelStatus) ([]Model, error) {
	var dtos []ModelResponse

	err := c.client.DoWithQuery(
		ctx,
		http.MethodGet,
		"/model/by-status",
		url.Values{
			"status": []string{string(status)},
		},
		nil,
		&dtos,
	)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	res := make([]Model, 0, len(dtos))
	for _, v := range dtos {
		res = append(res, *c.toDomainModel(&v))
	}

	return res, nil
}

func (c *HTTPModelClient) UpdateModel(ctx context.Context, id int, cmd UpdateModelCommand) (*Model, error) {
	req := UpdateModelRequest{
		Name:        cmd.Name,
		Version:     cmd.Version,
		ArtifactURI: cmd.ArtifactURI,
		Framework:   cmd.Framework,
		Status:      cmd.Status,
	}

	var dto ModelResponse
	err := c.client.Do(ctx, http.MethodPut, fmt.Sprintf("/model/%d", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainModel(&dto), nil
}

func (c *HTTPModelClient) ChangeStatus(ctx context.Context, id int, cmd ChangeStatusCommand) (*Model, error) {
	req := ChangeStatusRequest{
		Status: cmd.Status,
	}

	var dto ModelResponse
	err := c.client.Do(ctx, http.MethodPatch, fmt.Sprintf("/model/%d/status", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainModel(&dto), nil
}

func (c *HTTPModelClient) DeleteModel(ctx context.Context, id int) error {
	err := c.client.Do(ctx, http.MethodDelete, fmt.Sprintf("/model/%d", id), nil, nil)
	return c.client.MapErr(err, mapResponseError)
}

func (c *HTTPModelClient) DeployModel(ctx context.Context, cmd DeployCommand) (*DeploymentResult, error) {
	var dto DeployResponse
	path := fmt.Sprintf("/model/deploy/%d", cmd.ModelID)
	err := c.client.Do(ctx, http.MethodPost, path, nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	return &DeploymentResult{
		ID:     dto.DeploymentID,
		Status: ModelStatus(dto.Status),
	}, nil
}

func (c *HTTPModelClient) UndeployModel(ctx context.Context, cmd UnDeployCommand) (*DeploymentResult, error) {
	var dto UndeployResponse
	path := fmt.Sprintf("/model/undeploy/%d", cmd.ModelID)

	err := c.client.Do(ctx, http.MethodPost, path, nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	if !dto.Success {
		return nil, fmt.Errorf("undeploy failed: %s", dto.Message)
	}

	return &DeploymentResult{
		Status: ModelStaging,
	}, nil
}
