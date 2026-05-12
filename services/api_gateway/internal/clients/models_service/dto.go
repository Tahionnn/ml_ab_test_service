package modelsservice

import "github.com/google/uuid"

type CreateModelRequest struct {
	Name        string         `json:"name" validate:"required,min=2,max=50" example:"churn-model"`
	Version     string         `json:"version" validate:"required,is_version" example:"v1.0.0"`
	ArtifactURI string         `json:"artifact_uri" validate:"required,url" example:"https://models.internal/v2/res-net.zip"`
	Framework   ModelFramework `json:"framework" validate:"required" example:"identity"`
	Status      ModelStatus    `json:"status" validate:"omitempty,oneof=staging archived" example:"staging"`
}

type UpdateModelRequest struct {
	Name        *string         `json:"name,omitempty" validate:"omitempty,min=2,max=50" example:"churn-model"`
	Version     *string         `json:"version,omitempty" validate:"omitempty,is_version" example:"v2.0.0"`
	ArtifactURI *string         `json:"artifact_uri,omitempty" validate:"omitempty,url" example:"https://models.internal/v2/res-net.zip"`
	Framework   *ModelFramework `json:"framework,omitempty" example:"identity"`
	Status      *ModelStatus    `json:"status,omitempty" validate:"omitempty,oneof=staging archived" example:"staging"`
}

type SearchRequest struct {
	Q     string `form:"q" validate:"required"`
	Page  int    `form:"page,default=1" validate:"min=1"`
	Limit int    `form:"limit,default=20" validate:"min=1,max=100"`
}

type ChangeStatusRequest struct {
	Status ModelStatus `json:"status" validate:"oneof=staging archived" example:"staging"`
}

type ModelResponse struct {
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	Version         string         `json:"version"`
	ArtifactURI     string         `json:"artifact_uri"`
	ServingEndpoint *string        `json:"serving_endpoint"`
	Framework       ModelFramework `json:"framework"`
	Status          ModelStatus    `json:"status"`
	DeploymentID    *uuid.UUID     `json:"deployment_id"`
}

type SearchResponse struct {
	Total  int             `json:"total"`
	Models []ModelResponse `json:"models"`
}

type DeployResponse struct {
	DeploymentID uuid.UUID `json:"deployment_id"`
	Status       string    `json:"status"`
}

type UndeployResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
