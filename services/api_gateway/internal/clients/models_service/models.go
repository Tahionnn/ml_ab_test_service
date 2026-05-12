package modelsservice

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ModelStatus string

const (
	ModelStaging    ModelStatus = "staging"
	ModelPending    ModelStatus = "pending"
	ModelFailed     ModelStatus = "failed"
	ModelProduction ModelStatus = "production"
	ModelArchived   ModelStatus = "archived"
)

func (s ModelStatus) IsValid() bool {
	switch s {
	case ModelStaging, ModelPending, ModelFailed, ModelProduction, ModelArchived:
		return true
	}
	return false
}

func (s *ModelStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := ModelStatus(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid model status: %s", str)
	}

	*s = status
	return nil
}

type ModelFramework string

const (
	MLFlowFramework   ModelFramework = "mlflow"
	IdentityFramework ModelFramework = "identity"
	PyTorchFramework  ModelFramework = "pytorch"
	OnnxFramework     ModelFramework = "onnx"
)

func (s ModelFramework) IsValid() bool {
	switch s {
	case MLFlowFramework, IdentityFramework, PyTorchFramework, OnnxFramework:
		return true
	}
	return false
}

func (s *ModelFramework) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status := ModelFramework(str)
	if !status.IsValid() {
		return fmt.Errorf("invalid model framework: %s", str)
	}

	*s = status
	return nil
}

type Model struct {
	Name            string
	Version         string
	ArtifactUrl     string
	Framework       ModelFramework
	ID              *int
	ServingEndpoint *string
	Status          ModelStatus
	DeploymentID    *uuid.UUID
}

type DeploymentInfo struct {
	Endpoint     string
	DeploymentID uuid.UUID
	Metadata     map[string]any
}

type CreateModelCommand struct {
	Name        string
	Version     string
	ArtifactURI string
	Framework   ModelFramework
	Status      ModelStatus
}

type UpdateModelCommand struct {
	Name        *string
	Version     *string
	ArtifactURI *string
	Framework   *ModelFramework
	Status      *ModelStatus
}

type ChangeStatusCommand struct {
	ModelID int
	Status  ModelStatus
}

type SearchModelsQuery struct {
	Query string
	Page  int
	Limit int
}

type SearchResult struct {
	Total int
	Items []Model
}

type DeployCommand struct {
	ModelID int
}

type UnDeployCommand struct {
	ModelID int
}

type DeploymentResult struct {
	ID     uuid.UUID
	Status ModelStatus
}
