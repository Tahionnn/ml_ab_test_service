package domain

type VariantRequest struct {
	UserId       string
	ExperimentID string
	Attributes   map[string]string
}

type VariantResponse struct {
	VariantId        string
	ServingEndpoint  string
	IsControl        bool
	AssignmentReason string
}

type ExperimentConfig struct {
	ID         string
	Name       string
	Status     ExperimentStatus
	TrafficPct int
	Variants   []VariantConfig
}

type ExperimentStatus string

const (
	StatusDraft    ExperimentStatus = "draft"
	StatusRunning  ExperimentStatus = "running"
	StatusPaused   ExperimentStatus = "paused"
	StatusFinished ExperimentStatus = "finished"
)

type VariantConfig struct {
	ID              string
	Name            string
	ModelId         string
	ServingEndpoint string
	Weight          int
	IsControl       bool
	BucketStart     int
	BucketEnd       int
}

const TotalBuckets = 10_000
