from typing import Protocol, Any
import uuid

from domain.deployment.entities import DeploymentRecord
from domain.deployment.state import DeploymentStatus


class DeploymentStateRepository(Protocol):
    async def upsert(self, record: DeploymentRecord) -> None: ...

    async def get(self, deployment_id: str) -> DeploymentRecord | None: ...

    async def update_status(
        self,
        deployment_id: uuid.UUID,
        status: DeploymentStatus,
        *,
        error: str | None = None,
        **fields: Any,
    ) -> DeploymentRecord: ...

    async def create_if_not_exists(self, record: DeploymentRecord) -> bool: ...

    async def list_by_model_id(self, model_id: int) -> list[DeploymentRecord]: ...
