package experiment

import (
	"encoding/json"
	"time"
)

type CreateExperimentRequest struct {
	Name           string     `json:"name" validate:"required,min=1,max=255" example:"experiment"`
	Description    string     `json:"description,omitempty" validate:"max=1000" example:""`
	TrafficPercent int        `json:"traffic_percent" validate:"min=1,max=100" example:"100"`
	StartDate      time.Time  `json:"start_date" validate:"required" example:"2026-05-11T15:32:32.472Z"`
	EndDate        *time.Time `json:"end_date,omitempty" validate:"omitempty,gtfield=StartDate" example:"2026-06-11T15:32:32.472Z"`
}

func (r *CreateExperimentRequest) MarshalJSON() ([]byte, error) {
	type Alias CreateExperimentRequest
	if r.TrafficPercent == 0 {
		r.TrafficPercent = 100
	}
	return json.Marshal((*Alias)(r))
}

type UpdateExperimentRequest struct {
	Name           *string           `json:"name,omitempty" validate:"omitempty,min=1,max=255" example:"experiment"`
	Description    *string           `json:"description,omitempty" validate:"omitempty,max=1000" example:""`
	TrafficPercent *int              `json:"traffic_percent,omitempty" validate:"omitempty,min=1,max=100" example:"100"`
	Status         *ExperimentStatus `json:"status,omitempty" validate:"omitempty,oneof=draft running paused finished" example:"draft"`
	EndDate        *time.Time        `json:"end_date,omitempty" example:"2026-06-11T15:32:32.472Z"`
}

type CreateVariantRequest struct {
	VariantName   string `json:"variant_name" validate:"required,min=1,max=255" example:"variant"`
	ModelID       int    `json:"model_id" validate:"required,gt=0" example:"1"`
	TrafficWeight int    `json:"traffic_weight" validate:"required,min=0,max=100" example:"50"`
	IsControl     bool   `json:"is_control" example:"false"`
	Description   string `json:"description,omitempty" validate:"max=1000" example:""`
}

type UpdateVariantRequest struct {
	ModelID       *int  `json:"model_id,omitempty" validate:"omitempty,gt=0" example:"1"`
	TrafficWeight *int  `json:"traffic_weight,omitempty" validate:"omitempty,min=0,max=100" example:"50"`
	IsControl     *bool `json:"is_control,omitempty" example:"false"`
}

type AttachMetricRequest struct {
	MetricID int    `json:"metric_id" validate:"required,gt=0" example:"1"`
	Goal     string `json:"goal" validate:"max=1000" example:""`
}

type CreateMetricRequest struct {
	Name    string     `json:"name" validate:"required,min=1,max=255" example:"metric"`
	Type    MetricType `json:"type" validate:"required,oneof=counter gauge ratio" example:"ratio"`
	Formula string     `json:"formula" validate:"max=1000" example:""`
	Unit    Unit       `json:"unit" validate:"required,oneof=% ms count" example:"%"`
}

type UpdateMetricRequest struct {
	Name    *string     `json:"name,omitempty" validate:"omitempty,min=1,max=255" example:"metric"`
	Type    *MetricType `json:"type,omitempty" validate:"omitempty,oneof=counter gauge ratio" example:"ratio"`
	Formula *string     `json:"formula,omitempty" validate:"omitempty,max=1000" example:""`
	Unit    *Unit       `json:"unit,omitempty" validate:"omitempty,oneof=% ms count" example:"%"`
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

type MetricDTO struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	Type    MetricType `json:"type"`
	Formula string     `json:"formula"`
	Unit    Unit       `json:"unit"`
}

type ExperimentMetricDTO struct {
	MetricID     int    `json:"metric_id"`
	ExperimentID int    `json:"experiment_id"`
	Goal         string `json:"goal"`
}

type ExperimentDTO struct {
	ID             int                   `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Status         ExperimentStatus      `json:"status"`
	TrafficPercent int                   `json:"traffic_percent"`
	StartDate      time.Time             `json:"start_date"`
	EndDate        *time.Time            `json:"end_date,omitempty"`
	Variants       []VariantDTO          `json:"variants,omitempty"`
	Metrics        []ExperimentMetricDTO `json:"metrics,omitempty"`
}
