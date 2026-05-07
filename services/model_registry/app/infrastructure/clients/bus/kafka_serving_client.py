from typing import Any
import uuid

from domain.serving_client.client import ServingClient
from domain.serving_client.entities import DeploymentInfo

from core.broker import kafka_router

from infrastructure.clients.bus.schemas import (
    DeploymentRequestedEvent,
    UndeployRequestedEvent,
)

from infrastructure.clients.bus.topics import (
    DEPLOYMENT_REQUESTED,
    DEPLOYMENT_UNDEPLOY_REQUESTED,
)


class KafkaServingClient(ServingClient):  # type: ignore
    async def deploy(
        self,
        deployment_id: uuid.UUID,
        model_id: int,
        artifact_url: str,
        endpoint: str,
        model_type: str,
        metadata: dict[str, Any],
    ) -> DeploymentInfo:
        await kafka_router.broker.publish(
            DeploymentRequestedEvent(
                deployment_id=deployment_id,
                model_id=model_id,
                artifact_url=artifact_url,
                endpoint=endpoint,
                model_type=model_type,
                metadata=metadata,
            ),
            topic=DEPLOYMENT_REQUESTED,
        )

        return DeploymentInfo(
            deployment_id=deployment_id,
            endpoint=endpoint,
            metadata=metadata,
        )

    async def undeploy(self, deployment_id: str, model_id: int) -> None:
        await kafka_router.broker.publish(
            UndeployRequestedEvent(deployment_id=deployment_id, model_id=model_id),
            topic=DEPLOYMENT_UNDEPLOY_REQUESTED,
        )
