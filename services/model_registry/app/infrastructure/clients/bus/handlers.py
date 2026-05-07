from fastapi import Depends

from core.broker import kafka_router
from core.dependencies import get_model_service

from infrastructure.clients.bus.schemas import (
    DeploymentFailedEvent,
    DeploymentReadyEvent,
    UndeployedEvent,
)
from infrastructure.clients.bus.topics import (
    DEPLOYMENT_FAILED,
    DEPLOYMENT_READY,
    DEPLOYMENT_UNDEPLOYED,
)

from domain.model.service import ModelService


@kafka_router.subscriber(DEPLOYMENT_READY)  # type: ignore
async def handle_deployment_ready(
    event: DeploymentReadyEvent, service: ModelService = Depends(get_model_service)
) -> None:
    await service.on_deployment_ready(
        model_id=event.model_id, serving_endpoint=event.serving_url
    )


@kafka_router.subscriber(DEPLOYMENT_FAILED)  # type: ignore
async def handle_deployment_failed(
    event: DeploymentFailedEvent, service: ModelService = Depends(get_model_service)
) -> None:
    await service.on_deployment_failed(model_id=event.model_id)


@kafka_router.subscriber(DEPLOYMENT_UNDEPLOYED)  # type: ignore
async def handle_undeployed(
    event: UndeployedEvent, service: ModelService = Depends(get_model_service)
) -> None:
    await service.on_undeployment_ready(model_id=event.model_id)
