package experimentclient

import (
	"context"
	"time"

	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
)

type ExperimentClient interface {
	GetActiveExperiments(ctx context.Context) ([]ExperimentDTO, error)
}

type VariantDTO struct {
	ID            int    `json:"id"`
	ExperimentID  int    `json:"experiment_id"`
	Name          string `json:"name"`
	ModelID       int    `json:"model_id"`
	TrafficWeight int    `json:"traffic_weight"`
	IsControl     bool   `json:"is_control"`
	Description   string `json:"description"`
}

type ExperimentDTO struct {
	ID             int                     `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	Status         domain.ExperimentStatus `json:"status"`
	TrafficPercent int                     `json:"traffic_percent"`
	StartDate      *time.Time              `json:"start_date,omitempty"`
	EndDate        *time.Time              `json:"end_date,omitempty"`
	Variants       []VariantDTO            `json:"variants"`
}
