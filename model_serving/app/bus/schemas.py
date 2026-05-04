from typing import Any
import uuid

from pydantic import BaseModel, Field


class DeploymentRequestedEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int
    artifact_url: str
    endpoint: str
    model_type: str
    attempt: int = 0
    max_attempts: int = 3
    metadata: dict[str, Any] = Field(default_factory=dict)


class ArtifactDownloadedEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int
    local_path: str
    endpoint: str
    model_type: str
    attempt: int = 0
    max_attempts: int = 3
    metadata: dict[str, Any] = Field(default_factory=dict)


class DeploymentStartedEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int
    local_path: str
    endpoint: str
    model_type: str
    attempt: int = 0
    max_attempts: int = 3
    metadata: dict[str, Any] = Field(default_factory=dict)


class DeploymentReadyEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int
    serving_url: str
    deployment_name: str
    metadata: dict[str, Any] = Field(default_factory=dict)


class DeploymentFailedEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int
    stage: str
    error: str
    attempt: int = 0
    metadata: dict[str, Any] = Field(default_factory=dict)


class UndeployRequestedEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int


class UndeployedEvent(BaseModel):
    deployment_id: uuid.UUID
    model_id: int
