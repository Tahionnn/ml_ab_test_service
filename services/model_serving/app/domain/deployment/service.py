from typing import Any
import uuid

from loguru import logger

from domain.deployment.entities import DeploymentInfo, DeploymentRecord
from domain.deployment.state import DeploymentStatus
from domain.deployment.interface import ModelDeployer
from domain.deployment.state_repo import DeploymentStateRepository

from domain.artifact.interface import ArtifactSource


class DeploymentService:
    def __init__(
        self,
        deployer: ModelDeployer,
        state_repo: DeploymentStateRepository,
        artifact_source: ArtifactSource,
    ):
        self.deployer = deployer
        self.state_repo = state_repo
        self.artifact_source = artifact_source

    async def deploy(
        self,
        *,
        deployment_id: uuid.UUID,
        model_id: int,
        artifact_url: str,
        serving_endpoint: str,
        model_type: str,
    ) -> tuple[DeploymentInfo, DeploymentRecord]:
        info = await self.deployer.deploy(
            deployment_id=deployment_id,
            model_id=model_id,
            artifact_path=artifact_url,
            endpoint=serving_endpoint,
            model_type=model_type,
        )

        record = await self.state_repo.update_status(
            deployment_id,
            DeploymentStatus.READY,
            serving_url=info.endpoint,
            deployment_name=info.metadata.get("deployment_name"),
        )

        return info, record

    async def undeploy(self, deployment_id: uuid.UUID) -> None:
        await self.deployer.undeploy(deployment_id)
        await self.state_repo.update_status(deployment_id, DeploymentStatus.UNDEPLOYED)

    async def create_record(self, deployment_record: DeploymentRecord) -> None:
        await self.state_repo.upsert(deployment_record)

    async def get_or_create(
        self, record: DeploymentRecord
    ) -> tuple[DeploymentRecord, bool]:
        created = await self.state_repo.create_if_not_exists(record)
        existing = await self.state_repo.get(record.deployment_id)
        return existing, created

    async def exists(self, deployment_id: uuid.UUID) -> bool:
        return await self.state_repo.get(deployment_id) is not None

    async def get(self, deployment_id: uuid.UUID) -> DeploymentRecord | None:
        return await self.state_repo.get(deployment_id)

    async def download(
        self, deployment_id: uuid.UUID, artifact_url: str, attempt: int
    ) -> str:  # type: ignore[misc]
        await self.state_repo.update_status(
            deployment_id, DeploymentStatus.DOWNLOADING, attempt=attempt + 1
        )
        path = await self.artifact_source.download(artifact_url)
        return path

    async def update_status_downloaded(
        self, deployment_id: uuid.UUID, local_path: str
    ) -> None:
        await self.state_repo.update_status(
            deployment_id,
            DeploymentStatus.DEPLOYING,
            local_artifact_path=local_path,
        )

    async def update_status_failed(
        self, deployment_id: uuid.UUID, error: Exception | str
    ) -> None:
        await self.state_repo.update_status(
            deployment_id,
            DeploymentStatus.FAILED,
            error=str(error),
        )

    async def can_process_request(
        self, deployment_id: uuid.UUID, expected_statuses: set[DeploymentStatus]
    ) -> bool:
        record = await self.state_repo.get(deployment_id)
        if record is None:
            return False
        return record.status in expected_statuses

    async def transition_if_possible(
        self,
        deployment_id: uuid.UUID,
        target_status: DeploymentStatus,
        expected_current_statuses: set[DeploymentStatus],
        **update_fields: Any,
    ) -> DeploymentRecord | None:
        record = await self.state_repo.get(deployment_id)
        if record is None or record.status not in expected_current_statuses:
            logger.warning(
                f"Cannot transition deployment {deployment_id} from {record.status if record else 'None'} "
                f"to {target_status}, expected {expected_current_statuses}"
            )
            return None
        return await self.state_repo.update_status(
            deployment_id, target_status, **update_fields
        )
