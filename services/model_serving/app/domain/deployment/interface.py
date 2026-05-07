from typing import Protocol
import uuid

from domain.deployment.entities import DeploymentInfo


class ModelDeployer(Protocol):
    async def deploy(
        self,
        *,
        deployment_id: uuid.UUID,
        model_id: int,
        artifact_path: str,
        endpoint: str,
        model_type: str,
    ) -> DeploymentInfo: ...

    async def undeploy(self, deployment_id: uuid.UUID) -> None: ...

    async def scale(self, deployment_id: uuid.UUID, replicas: int) -> None: ...
