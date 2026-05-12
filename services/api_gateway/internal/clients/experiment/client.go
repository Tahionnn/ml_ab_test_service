package experiment

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/httpclient"
)

type ExperimentClient interface {
	// Experiments
	GetExperiment(ctx context.Context, id int) (*Experiment, error)
	//ListExperiments(ctx context.Context, page, pageSize int) ([]Experiment, error)
	CreateExperiment(ctx context.Context, cmd *CreateExperimentCommand) (*Experiment, error)
	UpdateExperiment(ctx context.Context, id int, cmd *UpdateExperimentCommand) (*Experiment, error)
	DeleteExperiment(ctx context.Context, id int) error

	// Experiment lifecycle
	StartExperiment(ctx context.Context, id int) (*Experiment, error)
	StopExperiment(ctx context.Context, id int) (*Experiment, error)
	ResumeExperiment(ctx context.Context, id int) (*Experiment, error)
	FinishExperiment(ctx context.Context, id int) (*Experiment, error)
	RestoreExperiment(ctx context.Context, id int) (*Experiment, error)
	//GetActiveExperiments(ctx context.Context) ([]Experiment, error)

	// Variants
	AddVariant(ctx context.Context, id int, cmd *CreateVariantCommand) (*Variant, error)
	GetVariant(ctx context.Context, id int) (*Variant, error)
	UpdateVariant(ctx context.Context, id int, cmd *UpdateVariantCommand) (*Variant, error)
	DeleteVariant(ctx context.Context, id int) error
	ListVariants(ctx context.Context, experimentID int) ([]Variant, error)

	// MetricsExperiment
	AttachMetric(ctx context.Context, experimentID int, cmd AttachMetricCommand) (*ExperimentMetric, error)
	ListExperimentMetrics(ctx context.Context, experimentID int) ([]ExperimentMetric, error)
	DetachMetric(ctx context.Context, experimentID, metricID int) error

	//Metrics
	GetMetric(ctx context.Context, id int) (*Metric, error)
	ListMetrics(ctx context.Context) ([]Metric, error)
	CreateMetric(ctx context.Context, cmd CreateMetricCommand) (*Metric, error)
	UpdateMetric(ctx context.Context, id int, cmd UpdateMetricCommand) (*Metric, error)
	DeleteMetric(ctx context.Context, id int) error
}

type HTTPExperimentClient struct {
	client *httpclient.Client
}

var _ ExperimentClient = (*HTTPExperimentClient)(nil)

type Config struct {
	BaseURL   string
	Timeout   time.Duration
	AuthToken string
}

func NewExperimentClient(cfg Config) *HTTPExperimentClient {
	return &HTTPExperimentClient{
		client: httpclient.New(cfg.BaseURL, cfg.Timeout, cfg.AuthToken),
	}
}

func (c *HTTPExperimentClient) toDomainExperiment(dto *ExperimentDTO) *Experiment {
	if dto == nil {
		return nil
	}

	variants := make([]Variant, len(dto.Variants))
	for i, v := range dto.Variants {
		variants[i] = *c.toDomainVariant(&v)
	}

	metrics := make([]ExperimentMetric, len(dto.Metrics))
	for i, m := range dto.Metrics {
		metrics[i] = ExperimentMetric{
			MetricID:     m.MetricID,
			ExperimentID: m.ExperimentID,
			Goal:         m.Goal,
		}
	}

	return &Experiment{
		ID:             dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Status:         dto.Status,
		TrafficPercent: dto.TrafficPercent,
		StartDate:      dto.StartDate,
		EndDate:        dto.EndDate,
		Variants:       variants,
		Metrics:        metrics,
	}
}

func (c *HTTPExperimentClient) toDomainVariant(dto *VariantDTO) *Variant {
	return &Variant{
		ID:            dto.ID,
		ExperimentID:  dto.ExperimentID,
		Name:          dto.Name,
		ModelID:       dto.ModelID,
		TrafficWeight: dto.TrafficWeight,
		IsControl:     dto.IsControl,
		Description:   dto.Description,
	}
}

func (c *HTTPExperimentClient) toDTOCreateExperiment(cmd *CreateExperimentCommand) *CreateExperimentRequest {
	req := &CreateExperimentRequest{
		Name:           cmd.Name,
		TrafficPercent: cmd.TrafficPercent,
		StartDate:      cmd.StartDate,
	}

	if cmd.Description != "" {
		req.Description = cmd.Description
	}
	if cmd.EndDate != nil {
		req.EndDate = cmd.EndDate
	}

	return req
}

func (c *HTTPExperimentClient) toDomainMetric(dto *MetricDTO) *Metric {
	return &Metric{
		ID:      dto.ID,
		Name:    dto.Name,
		Type:    MetricType(dto.Type),
		Formula: dto.Formula,
		Unit:    Unit(dto.Unit),
	}
}

func (c *HTTPExperimentClient) toDomainExperimentMetric(dto *ExperimentMetricDTO) *ExperimentMetric {
	return &ExperimentMetric{
		MetricID:     dto.MetricID,
		ExperimentID: dto.ExperimentID,
		Goal:         dto.Goal,
	}
}

func (c *HTTPExperimentClient) GetExperiment(ctx context.Context, id int) (*Experiment, error) {
	var dto ExperimentDTO
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/experiments/%d", id), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainExperiment(&dto), nil
}

func (c *HTTPExperimentClient) CreateExperiment(ctx context.Context, cmd *CreateExperimentCommand) (*Experiment, error) {
	req := CreateExperimentRequest{
		Name:           cmd.Name,
		Description:    cmd.Description,
		TrafficPercent: cmd.TrafficPercent,
		StartDate:      cmd.StartDate,
		EndDate:        cmd.EndDate,
	}

	var dto ExperimentDTO
	err := c.client.Do(ctx, http.MethodPost, "/experiments", req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainExperiment(&dto), nil
}

func (c *HTTPExperimentClient) UpdateExperiment(ctx context.Context, id int, cmd *UpdateExperimentCommand) (*Experiment, error) {
	req := UpdateExperimentRequest{
		Name:           cmd.Name,
		Description:    cmd.Description,
		TrafficPercent: cmd.TrafficPercent,
		Status:         cmd.Status,
		EndDate:        cmd.EndDate,
	}

	var dto ExperimentDTO
	err := c.client.Do(ctx, http.MethodPut, fmt.Sprintf("/experiments/%d", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainExperiment(&dto), nil
}

func (c *HTTPExperimentClient) DeleteExperiment(ctx context.Context, id int) error {
	err := c.client.Do(ctx, http.MethodDelete, fmt.Sprintf("/experiments/%d", id), nil, nil)
	return c.client.MapErr(err, mapResponseError)
}

// --- lifecycle ---

func (c *HTTPExperimentClient) StartExperiment(ctx context.Context, id int) (*Experiment, error) {
	return c.lifecycle(ctx, id, "start")
}

func (c *HTTPExperimentClient) StopExperiment(ctx context.Context, id int) (*Experiment, error) {
	return c.lifecycle(ctx, id, "pause")
}

func (c *HTTPExperimentClient) ResumeExperiment(ctx context.Context, id int) (*Experiment, error) {
	return c.lifecycle(ctx, id, "resume")
}

func (c *HTTPExperimentClient) FinishExperiment(ctx context.Context, id int) (*Experiment, error) {
	return c.lifecycle(ctx, id, "finish")
}

func (c *HTTPExperimentClient) RestoreExperiment(ctx context.Context, id int) (*Experiment, error) {
	return c.lifecycle(ctx, id, "restore")
}

func (c *HTTPExperimentClient) lifecycle(ctx context.Context, id int, action string) (*Experiment, error) {
	var dto ExperimentDTO
	err := c.client.Do(ctx, http.MethodPost, fmt.Sprintf("/experiments/%d/%s", id, action), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainExperiment(&dto), nil
}

// --- variants ---

func (c *HTTPExperimentClient) AddVariant(ctx context.Context, id int, cmd *CreateVariantCommand) (*Variant, error) {
	req := CreateVariantRequest{
		VariantName:   cmd.Name,
		ModelID:       cmd.ModelID,
		TrafficWeight: cmd.TrafficWeight,
		IsControl:     cmd.IsControl,
		Description:   cmd.Description,
	}

	var dto VariantDTO
	err := c.client.Do(ctx, http.MethodPost, fmt.Sprintf("/experiments/%d/variants", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainVariant(&dto), nil
}

func (c *HTTPExperimentClient) GetVariant(ctx context.Context, id int) (*Variant, error) {
	var dto VariantDTO
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/variants/%d", id), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainVariant(&dto), nil
}

func (c *HTTPExperimentClient) UpdateVariant(ctx context.Context, id int, cmd *UpdateVariantCommand) (*Variant, error) {
	req := UpdateVariantRequest{
		ModelID:       cmd.ModelID,
		TrafficWeight: cmd.TrafficWeight,
		IsControl:     cmd.IsControl,
	}

	var dto VariantDTO
	err := c.client.Do(ctx, http.MethodPut, fmt.Sprintf("/variants/%d", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainVariant(&dto), nil
}

func (c *HTTPExperimentClient) DeleteVariant(ctx context.Context, id int) error {
	err := c.client.Do(ctx, http.MethodDelete, fmt.Sprintf("/variants/%d", id), nil, nil)
	return c.client.MapErr(err, mapResponseError)
}

func (c *HTTPExperimentClient) ListVariants(ctx context.Context, experimentID int) ([]Variant, error) {
	var dtos []VariantDTO
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/experiments/%d/variants", experimentID), nil, &dtos)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	res := make([]Variant, 0, len(dtos))
	for _, v := range dtos {
		res = append(res, *c.toDomainVariant(&v))
	}
	return res, nil
}

// --- experiment metrics ---

func (c *HTTPExperimentClient) AttachMetric(ctx context.Context, experimentID int, cmd AttachMetricCommand) (*ExperimentMetric, error) {
	req := AttachMetricRequest{
		MetricID: cmd.MetricID,
		Goal:     cmd.Goal,
	}

	var dto ExperimentMetricDTO
	err := c.client.Do(ctx, http.MethodPost, fmt.Sprintf("/experiments/%d/metrics", experimentID), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainExperimentMetric(&dto), nil
}

func (c *HTTPExperimentClient) ListExperimentMetrics(ctx context.Context, experimentID int) ([]ExperimentMetric, error) {
	var dtos []ExperimentMetricDTO
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/experiments/%d/metrics", experimentID), nil, &dtos)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	res := make([]ExperimentMetric, 0, len(dtos))
	for _, m := range dtos {
		res = append(res, *c.toDomainExperimentMetric(&m))
	}
	return res, nil
}

func (c *HTTPExperimentClient) DetachMetric(ctx context.Context, experimentID, metricID int) error {
	err := c.client.Do(ctx, http.MethodDelete,
		fmt.Sprintf("/experiments/%d/metrics/%d", experimentID, metricID),
		nil,
		nil,
	)
	return c.client.MapErr(err, mapResponseError)
}

// --- metrics ---

func (c *HTTPExperimentClient) GetMetric(ctx context.Context, id int) (*Metric, error) {
	var dto MetricDTO
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/metrics/%d", id), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainMetric(&dto), nil
}

func (c *HTTPExperimentClient) ListMetrics(ctx context.Context) ([]Metric, error) {
	var dtos []MetricDTO
	err := c.client.Do(ctx, http.MethodGet, "/metrics", nil, &dtos)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	res := make([]Metric, 0, len(dtos))
	for _, m := range dtos {
		res = append(res, *c.toDomainMetric(&m))
	}
	return res, nil
}

func (c *HTTPExperimentClient) CreateMetric(ctx context.Context, cmd CreateMetricCommand) (*Metric, error) {
	req := CreateMetricRequest{
		Name:    cmd.Name,
		Type:    cmd.Type,
		Formula: cmd.Formula,
		Unit:    cmd.Unit,
	}

	var dto MetricDTO
	err := c.client.Do(ctx, http.MethodPost, "/metrics", req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainMetric(&dto), nil
}

func (c *HTTPExperimentClient) UpdateMetric(ctx context.Context, id int, cmd UpdateMetricCommand) (*Metric, error) {
	req := UpdateMetricRequest{
		Name:    cmd.Name,
		Type:    cmd.Type,
		Formula: cmd.Formula,
		Unit:    cmd.Unit,
	}

	var dto MetricDTO
	err := c.client.Do(ctx, http.MethodPut, fmt.Sprintf("/metrics/%d", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainMetric(&dto), nil
}

func (c *HTTPExperimentClient) DeleteMetric(ctx context.Context, id int) error {
	err := c.client.Do(ctx, http.MethodDelete, fmt.Sprintf("/metrics/%d", id), nil, nil)
	return c.client.MapErr(err, mapResponseError)
}
