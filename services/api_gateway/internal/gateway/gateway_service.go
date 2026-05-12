package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/experiment"
	modelsservice "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/models_service"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/splitter"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/apierror"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/broker"
	"go.uber.org/zap"
)

type Service struct {
	users       user.UserClient
	experiments experiment.ExperimentClient
	models      modelsservice.ModelClient
	splitter    splitter.SplitterService
	events      broker.EventProducer
	httpClient  *http.Client
	lg          *zap.Logger
}

func NewService(
	users user.UserClient,
	experiments experiment.ExperimentClient,
	models modelsservice.ModelClient,
	splitter splitter.SplitterService,
	events broker.EventProducer,
	lg *zap.Logger,
) *Service {
	return &Service{
		users:       users,
		experiments: experiments,
		models:      models,
		splitter:    splitter,
		events:      events,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		lg:          lg,
	}
}

// ─── AUTH ────────────────────────────────────────────────────────────────────

func (s *Service) Register(ctx context.Context, cmd user.RegisterCommand) (*user.User, error) {
	return s.users.Register(ctx, cmd)
}

func (s *Service) Login(ctx context.Context, username, password string) (*user.Token, error) {
	return s.users.Login(ctx, username, password)
}

func (s *Service) VerifyToken(ctx context.Context, token string) (*user.InternalUser, error) {
	return s.users.VerifyToken(ctx, token)
}

// ─── USERS ───────────────────────────────────────────────────────────────────

func (s *Service) GetUser(ctx context.Context, id int) (*user.User, error) {
	return s.users.GetUser(ctx, id)
}

func (s *Service) DeleteUser(ctx context.Context, id int) error {
	return s.users.DeleteUser(ctx, id)
}

func (s *Service) UpdateUserRole(ctx context.Context, id int, cmd user.UserRoleUpdateCommand) (*user.User, error) {
	return s.users.UpdateUserRole(ctx, id, cmd)
}

// ─── EXPERIMENTS ─────────────────────────────────────────────────────────────

func (s *Service) GetExperiment(ctx context.Context, id int) (*experiment.Experiment, error) {
	return s.experiments.GetExperiment(ctx, id)
}

func (s *Service) CreateExperiment(ctx context.Context, cmd *experiment.CreateExperimentCommand) (*experiment.Experiment, error) {
	return s.experiments.CreateExperiment(ctx, cmd)
}

func (s *Service) UpdateExperiment(ctx context.Context, id int, cmd *experiment.UpdateExperimentCommand) (*experiment.Experiment, error) {
	return s.experiments.UpdateExperiment(ctx, id, cmd)
}

func (s *Service) DeleteExperiment(ctx context.Context, id int) error {
	return s.experiments.DeleteExperiment(ctx, id)
}

func (s *Service) StartExperiment(ctx context.Context, id int) (*experiment.Experiment, error) {
	exp, err := s.experiments.GetExperiment(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.validateVariantModels(ctx, exp.Variants); err != nil {
		return nil, err
	}

	return s.experiments.StartExperiment(ctx, id)
}

func (s *Service) validateVariantModels(ctx context.Context, variants []experiment.Variant) error {
	for _, v := range variants {
		model, err := s.models.GetModel(ctx, v.ModelID)
		if err != nil {
			return fmt.Errorf("%w: variant '%s' references model %d which doesn't exist",
				ErrModelNotFound, v.Name, v.ModelID)
		}
		if model.Status != modelsservice.ModelProduction {
			return fmt.Errorf("%w: variant '%s' uses model '%s' (status: %s)",
				ErrModelNotProduction, v.Name, model.Name, model.Status)
		}
	}
	return nil
}

func (s *Service) StopExperiment(ctx context.Context, id int) (*experiment.Experiment, error) {
	return s.experiments.StopExperiment(ctx, id)
}

func (s *Service) ResumeExperiment(ctx context.Context, id int) (*experiment.Experiment, error) {
	return s.experiments.ResumeExperiment(ctx, id)
}

func (s *Service) FinishExperiment(ctx context.Context, id int) (*experiment.Experiment, error) {
	return s.experiments.FinishExperiment(ctx, id)
}

func (s *Service) RestoreExperiment(ctx context.Context, id int) (*experiment.Experiment, error) {
	return s.experiments.RestoreExperiment(ctx, id)
}

// ─── VARIANTS ────────────────────────────────────────────────────────────────

func (s *Service) AddVariant(ctx context.Context, experimentID int, cmd *experiment.CreateVariantCommand) (*experiment.Variant, error) {
	model, err := s.models.GetModel(ctx, cmd.ModelID)
	if err != nil {
		return nil, fmt.Errorf("%w: model %d not found in registry", ErrModelNotFound, cmd.ModelID)
	}

	exp, err := s.experiments.GetExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	if exp.Status != experiment.StatusDraft && model.Status != modelsservice.ModelProduction {
		return nil, fmt.Errorf("%w: experiment is %s, model must be production (got %s)",
			ErrModelNotProduction, exp.Status, model.Status)
	}

	return s.experiments.AddVariant(ctx, experimentID, cmd)
}

func (s *Service) GetVariant(ctx context.Context, id int) (*experiment.Variant, error) {
	return s.experiments.GetVariant(ctx, id)
}

func (s *Service) UpdateVariant(ctx context.Context, id int, cmd *experiment.UpdateVariantCommand) (*experiment.Variant, error) {
	if cmd.ModelID != nil {
		model, err := s.models.GetModel(ctx, *cmd.ModelID)
		if err != nil {
			return nil, fmt.Errorf("%w: model %d not found", ErrModelNotFound, *cmd.ModelID)
		}
		existing, _ := s.experiments.GetVariant(ctx, id)
		if existing != nil {
			exp, _ := s.experiments.GetExperiment(ctx, existing.ExperimentID)
			if exp != nil && exp.Status != experiment.StatusDraft && model.Status != modelsservice.ModelProduction {
				return nil, fmt.Errorf("%w: cannot change to non-production model in active experiment", ErrModelNotProduction)
			}
		}
	}

	return s.experiments.UpdateVariant(ctx, id, cmd)
}

func (s *Service) DeleteVariant(ctx context.Context, id int) error {
	return s.experiments.DeleteVariant(ctx, id)
}

func (s *Service) ListVariants(ctx context.Context, experimentID int) ([]experiment.Variant, error) {
	return s.experiments.ListVariants(ctx, experimentID)
}

// ─── METRICS ─────────────────────────────────────────────────────────────────

func (s *Service) ListMetrics(ctx context.Context) ([]experiment.Metric, error) {
	return s.experiments.ListMetrics(ctx)
}

func (s *Service) GetMetric(ctx context.Context, id int) (*experiment.Metric, error) {
	return s.experiments.GetMetric(ctx, id)
}

func (s *Service) CreateMetric(ctx context.Context, cmd experiment.CreateMetricCommand) (*experiment.Metric, error) {
	return s.experiments.CreateMetric(ctx, cmd)
}

func (s *Service) UpdateMetric(ctx context.Context, id int, cmd experiment.UpdateMetricCommand) (*experiment.Metric, error) {
	return s.experiments.UpdateMetric(ctx, id, cmd)
}

func (s *Service) DeleteMetric(ctx context.Context, id int) error {
	return s.experiments.DeleteMetric(ctx, id)
}

func (s *Service) AttachMetric(ctx context.Context, expID int, cmd experiment.AttachMetricCommand) (*experiment.ExperimentMetric, error) {
	return s.experiments.AttachMetric(ctx, expID, cmd)
}

func (s *Service) ListExperimentMetrics(ctx context.Context, expID int) ([]experiment.ExperimentMetric, error) {
	return s.experiments.ListExperimentMetrics(ctx, expID)
}

func (s *Service) DetachMetric(ctx context.Context, expID, metricID int) error {
	return s.experiments.DetachMetric(ctx, expID, metricID)
}

// ─── MODELS ──────────────────────────────────────────────────────────────────

func (s *Service) CreateModel(ctx context.Context, cmd modelsservice.CreateModelCommand) (*modelsservice.Model, error) {
	return s.models.CreateModel(ctx, cmd)
}

func (s *Service) ListModels(ctx context.Context) ([]modelsservice.Model, error) {
	return s.models.ListModels(ctx)
}

func (s *Service) GetModel(ctx context.Context, id int) (*modelsservice.Model, error) {
	return s.models.GetModel(ctx, id)
}

func (s *Service) UpdateModel(ctx context.Context, id int, cmd modelsservice.UpdateModelCommand) (*modelsservice.Model, error) {
	return s.models.UpdateModel(ctx, id, cmd)
}

func (s *Service) DeleteModel(ctx context.Context, id int) error {
	return s.models.DeleteModel(ctx, id)
}

func (s *Service) DeployModel(ctx context.Context, id int) (*modelsservice.DeploymentResult, error) {
	return s.models.DeployModel(ctx, modelsservice.DeployCommand{ModelID: id})
}

func (s *Service) UndeployModel(ctx context.Context, id int) (*modelsservice.DeploymentResult, error) {
	return s.models.UndeployModel(ctx, modelsservice.UnDeployCommand{ModelID: id})
}

func (s *Service) ChangeModelStatus(ctx context.Context, id int, cmd modelsservice.ChangeStatusCommand) (*modelsservice.Model, error) {
	return s.models.ChangeStatus(ctx, id, cmd)
}

// ─── PREDICT ─────────────────────────────────────────────────────────────────

type PredictRequest struct {
	UserID       string         `json:"user_id"`
	ExperimentID string         `json:"experiment_id"`
	Payload      map[string]any `json:"payload"`
}

type PredictResponse struct {
	VariantID        string         `json:"variant_id"`
	IsControl        bool           `json:"is_control"`
	AssignmentReason string         `json:"assignment_reason"`
	Prediction       map[string]any `json:"prediction"`
}

func (s *Service) Predict(ctx context.Context, req PredictRequest) (*PredictResponse, error) {
	variant, err := s.splitter.GetVariant(ctx, req.UserID, req.ExperimentID, nil)
	if err != nil {
		// Splitter speaks gRPC, so errors arrive as sentinel values — wrap for handleErr.
		return nil, fmt.Errorf("%w: %s", ErrSplitterUnavailable, err)
	}

	if variant.ServingEndpoint == "" {
		return nil, fmt.Errorf("%w: variant '%s' has no serving endpoint", ErrServingUnavailable, variant.VariantID)
	}

	prediction, err := s.callServingEndpoint(ctx, variant.ServingEndpoint, req.Payload)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrServingUnavailable, err)
	}

	go func() {
		_ = s.events.PublishPredictionEvent(context.Background(), broker.PredictionEvent{
			UserID:       req.UserID,
			ExperimentID: req.ExperimentID,
			VariantID:    variant.VariantID,
			IsControl:    variant.IsControl,
			Timestamp:    time.Now(),
		})
	}()

	return &PredictResponse{
		VariantID:        variant.VariantID,
		IsControl:        variant.IsControl,
		AssignmentReason: variant.AssignmentReason,
		Prediction:       prediction,
	}, nil
}

func (s *Service) callServingEndpoint(ctx context.Context, endpoint string, payload map[string]any) (map[string]any, error) {
	body, err := json.Marshal(map[string]any{"features": payload})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("serving returned %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// ─── SIMULATE ────────────────────────────────────────────────────────────────

type SimulateRequest struct {
	ExperimentID string         `json:"experiment_id"`
	UserIDs      []string       `json:"user_ids"`
	Payload      map[string]any `json:"payload"`
}

type SimulateResult struct {
	UserID    string `json:"user_id"`
	VariantID string `json:"variant_id"`
	IsControl bool   `json:"is_control"`
	Error     string `json:"error,omitempty"`
}

func (s *Service) Simulate(ctx context.Context, req SimulateRequest) ([]SimulateResult, error) {
	results := make([]SimulateResult, 0, len(req.UserIDs))

	for _, userID := range req.UserIDs {
		variant, err := s.splitter.GetVariant(ctx, userID, req.ExperimentID, nil)
		if err != nil {
			results = append(results, SimulateResult{UserID: userID, Error: err.Error()})
			continue
		}

		_ = s.events.PublishPredictionEvent(ctx, broker.PredictionEvent{
			UserID:       userID,
			ExperimentID: req.ExperimentID,
			VariantID:    variant.VariantID,
			IsControl:    variant.IsControl,
			Timestamp:    time.Now(),
		})

		results = append(results, SimulateResult{
			UserID:    userID,
			VariantID: variant.VariantID,
			IsControl: variant.IsControl,
		})
	}

	return results, nil
}

// ─── HEALTH ──────────────────────────────────────────────────────────────────

type HealthStatus struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

func (s *Service) CheckHealth(ctx context.Context) []HealthStatus {
	type result struct {
		name string
		err  error
	}

	checks := []struct {
		name string
		fn   func() error
	}{
		{"user_service", func() error { _, err := s.users.GetUser(ctx, 0); return ignoreNotFound(err) }},
		{"experiment_manager", func() error { _, err := s.experiments.GetExperiment(ctx, 0); return ignoreNotFound(err) }},
		{"model_registry", func() error { _, err := s.models.GetModel(ctx, 0); return ignoreNotFound(err) }},
	}

	statuses := make([]HealthStatus, 0, len(checks))
	ch := make(chan result, len(checks))

	for _, c := range checks {
		go func(name string, fn func() error) {
			ch <- result{name: name, err: fn()}
		}(c.name, c.fn)
	}

	for range checks {
		r := <-ch
		hs := HealthStatus{Service: r.name, Status: "ok"}
		if r.err != nil {
			hs.Status = "unavailable"
			hs.Error = r.err.Error()
		}
		statuses = append(statuses, hs)
	}
	return statuses
}

func ignoreNotFound(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *apierror.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return nil
	}

	if errors.Is(err, ErrModelNotFound) {
		return nil
	}
	return err
}
