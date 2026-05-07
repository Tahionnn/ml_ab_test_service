from dataclasses import dataclass, field
from datetime import datetime, UTC
from typing import Any
import uuid

from domain.deployment.state import DeploymentStatus


@dataclass
class DeploymentInfo:
    endpoint: str
    deployment_id: uuid.UUID
    metadata: dict[str, Any] = field(default_factory=dict)


@dataclass(slots=True)
class DeploymentRecord:
    deployment_id: uuid.UUID
    model_id: int
    artifact_url: str
    endpoint: str
    model_type: str
    status: DeploymentStatus = DeploymentStatus.REQUESTED
    attempt: int = 0
    error: str | None = None
    deployment_name: str | None = None
    serving_url: str | None = None
    local_artifact_path: str | None = None
    created_at: datetime = field(default_factory=lambda: datetime.now(UTC))
    updated_at: datetime = field(default_factory=lambda: datetime.now(UTC))
    metadata: dict[str, Any] = field(default_factory=dict)

    def touch(self) -> None:
        self.updated_at = datetime.now(UTC)
