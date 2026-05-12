package experiment

import (
	"encoding/json"
	"fmt"
	"time"
)

type ExperimentStatus string

const (
	StatusDraft    ExperimentStatus = "draft"
	StatusRunning  ExperimentStatus = "running"
	StatusPaused   ExperimentStatus = "paused"
	StatusFinished ExperimentStatus = "finished"
)

func (s ExperimentStatus) IsValid() bool {
	switch s {
	case StatusDraft, StatusRunning, StatusPaused, StatusFinished:
		return true
	}
	return false
}

func (s *ExperimentStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := ExperimentStatus(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid experiment status: %s", str)
	}

	*s = status
	return nil
}

type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeRatio   MetricType = "ratio"
)

func (t MetricType) IsValid() bool {
	switch t {
	case MetricTypeCounter, MetricTypeGauge, MetricTypeRatio:
		return true
	}
	return false
}

func (t *MetricType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := MetricType(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid metric type: %s", str)
	}

	*t = status
	return nil
}

type Unit string

const (
	UnitPercent Unit = "%"
	UnitSeconds Unit = "ms"
	UnitBytes   Unit = "count"
)

func (s Unit) IsValid() bool {
	switch s {
	case UnitPercent, UnitSeconds, UnitBytes:
		return true
	}
	return false
}

func (s *Unit) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := Unit(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid unit: %s", str)
	}

	*s = status
	return nil
}

type Experiment struct {
	ID             int
	Name           string
	Description    string
	Status         ExperimentStatus
	TrafficPercent int
	StartDate      time.Time
	EndDate        *time.Time
	Variants       []Variant
	Metrics        []ExperimentMetric
}

type Variant struct {
	ID            int
	ExperimentID  int
	Name          string
	ModelID       int
	TrafficWeight int
	IsControl     bool
	Description   string
}

type Metric struct {
	ID      int
	Name    string
	Type    MetricType
	Formula string
	Unit    Unit
}

type ExperimentMetric struct {
	MetricID     int
	ExperimentID int
	Goal         string
}

type CreateExperimentCommand struct {
	Name           string
	Description    string
	TrafficPercent int
	StartDate      time.Time
	EndDate        *time.Time
}

type UpdateExperimentCommand struct {
	Name           *string
	Description    *string
	TrafficPercent *int
	Status         *ExperimentStatus
	EndDate        *time.Time
}

type CreateVariantCommand struct {
	ExperimentID  int
	Name          string
	ModelID       int
	TrafficWeight int
	IsControl     bool
	Description   string
}

type UpdateVariantCommand struct {
	ModelID       *int
	TrafficWeight *int
	IsControl     *bool
}

type AttachMetricCommand struct {
	MetricID int
	Goal     string
}

type CreateMetricCommand struct {
	Name    string
	Type    MetricType
	Formula string
	Unit    Unit
}

type UpdateMetricCommand struct {
	Name    *string
	Type    *MetricType
	Formula *string
	Unit    *Unit
}
