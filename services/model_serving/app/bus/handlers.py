from typing import TypeVar, Any, cast
from collections.abc import Callable

from faststream import Depends
from loguru import logger
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type,
    before_sleep_log,
)

from bus.broker import broker
from bus.schemas import (
    ArtifactDownloadedEvent,
    DeploymentFailedEvent,
    DeploymentReadyEvent,
    DeploymentRequestedEvent,
    DeploymentStartedEvent,
    UndeployRequestedEvent,
    UndeployedEvent,
)
from bus.topics import (
    ARTIFACT_DOWNLOADED,
    DEPLOYMENT_FAILED,
    DEPLOYMENT_READY,
    DEPLOYMENT_REQUESTED,
    DEPLOYMENT_STARTED,
    DEPLOYMENT_UNDEPLOYED,
    DEPLOYMENT_UNDEPLOY_REQUESTED,
)

from domain.deployment.entities import (
    DeploymentInfo,
    DeploymentRecord,
    DeploymentStatus,
)
from domain.deployment.service import DeploymentService

from core.config import settings
from core.dependencies import get_deployment_service


F = TypeVar("F", bound=Callable[..., Any])


def deployment_retry_policy() -> Callable[[F], F]:
    decorator = retry(
        stop=stop_after_attempt(settings.deployment_retry_attempts),
        wait=wait_exponential(
            multiplier=settings.deployment_retry_base_delay_sec, max=60
        ),
        retry=retry_if_exception_type(Exception),
        before_sleep=before_sleep_log(logger, 30),
        reraise=True,
    )
    return cast(Callable[[F], F], decorator)


async def _fail(
    event: DeploymentRequestedEvent | ArtifactDownloadedEvent | DeploymentStartedEvent,
    stage: str,
    error: Exception | str,
    service: DeploymentService,
) -> None:
    try:
        await service.update_status_failed(
            event.deployment_id,
            error=error,
        )
    except Exception:
        logger.exception("Failed to update status to FAILED")

    try:
        await broker.publish(
            DeploymentFailedEvent(
                deployment_id=event.deployment_id,
                model_id=event.model_id,
                stage=stage,
                error=str(error),
                attempt=getattr(event, "attempt", 0),
                metadata=getattr(event, "metadata", {}),
            ),
            topic=DEPLOYMENT_FAILED,
        )
    except Exception:
        logger.exception("Failed to publish DeploymentFailedEvent")


@broker.subscriber(DEPLOYMENT_REQUESTED)  # type: ignore
async def handle_deployment_request(
    event: DeploymentRequestedEvent,
    service: DeploymentService = Depends(get_deployment_service),
) -> None:

    record = DeploymentRecord(
        deployment_id=event.deployment_id,
        model_id=event.model_id,
        artifact_url=event.artifact_url,
        endpoint=event.endpoint,
        model_type=event.model_type,
        attempt=event.attempt,
        metadata=event.metadata,
    )

    _, created = await service.get_or_create(record)

    if not created:
        if not await service.can_process_request(
            event.deployment_id,
            {DeploymentStatus.REQUESTED, DeploymentStatus.FAILED},
        ):
            logger.info(f"Deployment {event.deployment_id} already processed, skipping")
            return

        await service.transition_if_possible(
            event.deployment_id,
            DeploymentStatus.REQUESTED,
            {DeploymentStatus.FAILED, DeploymentStatus.REQUESTED},
        )

    @deployment_retry_policy()
    async def _download_with_retry() -> str:
        local_path = await service.download(
            event.deployment_id,
            event.artifact_url,
            event.attempt,
        )
        return local_path

    try:
        local_path = await _download_with_retry()
    except Exception as e:
        await _fail(event, "download", e, service)
        return

    try:

        await broker.publish(
            ArtifactDownloadedEvent(
                deployment_id=event.deployment_id,
                model_id=event.model_id,
                local_path=local_path,
                endpoint=event.endpoint,
                model_type=event.model_type,
                attempt=event.attempt,
                max_attempts=event.max_attempts,
                metadata=event.metadata,
            ),
            topic=ARTIFACT_DOWNLOADED,
        )
    except Exception as e:
        await _fail(event, "download", e, service)


@broker.subscriber(ARTIFACT_DOWNLOADED)  # type: ignore
async def handle_artifact_downloaded(
    event: ArtifactDownloadedEvent,
    service: DeploymentService = Depends(get_deployment_service),
) -> None:

    if not await service.can_process_request(
        event.deployment_id, {DeploymentStatus.DOWNLOADING}
    ):
        return

    try:
        await service.update_status_downloaded(
            event.deployment_id,
            event.local_path,
        )
    except Exception as e:
        await _fail(event, "download", e, service)
        return

    @deployment_retry_policy()
    async def _publish_with_retry() -> None:
        await broker.publish(
            DeploymentStartedEvent(
                deployment_id=event.deployment_id,
                model_id=event.model_id,
                local_path=event.local_path,
                endpoint=event.endpoint,
                model_type=event.model_type,
                attempt=event.attempt,
                max_attempts=event.max_attempts,
                metadata=event.metadata,
            ),
            topic=DEPLOYMENT_STARTED,
        )

    try:
        await _publish_with_retry()
    except Exception as e:
        await _fail(event, "transition_to_deploy", e, service)


@broker.subscriber(DEPLOYMENT_STARTED)  # type: ignore
async def handle_deployment_started(
    event: DeploymentStartedEvent,
    service: DeploymentService = Depends(get_deployment_service),
) -> None:

    record = await service.get(event.deployment_id)

    if record is None:
        return

    if record.status == DeploymentStatus.READY:
        logger.info(f"{event.deployment_id} already READY, skipping duplicate")
        return

    if record.status != DeploymentStatus.DEPLOYING:
        return

    @deployment_retry_policy()
    async def _deploy_with_retry() -> tuple[DeploymentInfo, DeploymentRecord]:
        info, record = await service.deploy(
            deployment_id=event.deployment_id,
            model_id=event.model_id,
            artifact_url=event.local_path,
            serving_endpoint=event.endpoint,
            model_type=event.model_type,
        )

        return info, record

    try:
        info, record = await _deploy_with_retry()
    except Exception as e:
        await _fail(event, "download", e, service)
        return

    try:
        await broker.publish(
            DeploymentReadyEvent(
                deployment_id=event.deployment_id,
                model_id=event.model_id,
                serving_url=record.serving_url or info.endpoint,
                deployment_name=info.metadata.get(
                    "deployment_name", f"model_{event.deployment_id}"
                ),
                metadata=info.metadata,
            ),
            topic=DEPLOYMENT_READY,
        )
    except Exception as e:
        await _fail(event, "deploy", e, service)


@broker.subscriber(DEPLOYMENT_UNDEPLOY_REQUESTED)  # type: ignore
async def handle_undeploy_requested(
    event: UndeployRequestedEvent,
    service: DeploymentService = Depends(get_deployment_service),
) -> None:
    @deployment_retry_policy()
    async def _undeploy_with_retry() -> None:
        await service.undeploy(event.deployment_id)

    try:
        await _undeploy_with_retry()
    except Exception as e:
        await _fail(event, "download", e, service)
        return

    try:
        await broker.publish(
            UndeployedEvent(deployment_id=event.deployment_id, model_id=event.model_id),
            topic=DEPLOYMENT_UNDEPLOYED,
        )
    except Exception as e:
        await _fail(event, "undeploy", e, service)
