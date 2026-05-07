from redis.asyncio import Redis

from core.config import settings

from domain.deployment.service import DeploymentService
from domain.artifact.interface import ArtifactSource

from infrastructure.deployers.ray.ray_deployer import RayDeployer
from infrastructure.state_repo.redis_state_repo import RedisDeploymentStateRepository

from infrastructure.artifact.artifact_manager import ArtifactManager
from infrastructure.artifact.artifact_sources.mock_sourse import MockArtifactSource


def get_artifact_manager() -> ArtifactSource:
    sources = {
        "mock": MockArtifactSource(),
    }
    return ArtifactManager(sources=sources, default_source=MockArtifactSource())


def get_deployment_service() -> DeploymentService:
    deployer = RayDeployer(settings.RAY_SERVE_ADDRESS)

    redis = Redis.from_url(settings.REDIS_URL, decode_responses=True)
    repo = RedisDeploymentStateRepository(redis)

    artifact_source = get_artifact_manager()
    return DeploymentService(
        deployer=deployer, state_repo=repo, artifact_source=artifact_source
    )
