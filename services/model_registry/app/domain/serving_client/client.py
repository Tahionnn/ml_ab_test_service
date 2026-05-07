from typing import Protocol, Any
import uuid

from domain.serving_client.entities import DeploymentInfo


class ServingClient(Protocol):
    async def deploy(
        self,
        deployment_id: uuid.UUID,
        model_id: int,
        artifact_utl: str,
        endpoint: str,
        model_type: str,
        metadata: dict[str, Any],
    ) -> DeploymentInfo: ...

    async def undeploy(self, deployment_id: str, model_id: int) -> None: ...
