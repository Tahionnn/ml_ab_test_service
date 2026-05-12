import uuid
from loguru import logger

import ray
from ray import serve
from ray.serve.exceptions import RayServeException

from domain.deployment.entities import DeploymentInfo
from domain.deployment.interface import ModelDeployer
from domain.deployment.exceptions import UnsupportedModelType

from infrastructure.deployers.ray.model_type import _MODEL_TYPE

from core.config import settings


class RayDeployer(ModelDeployer):  # type: ignore
    _initialized = False

    def __init__(self, address: str):
        self.base_url = settings.SERVE_URL

        if not RayDeployer._initialized:
            if not ray.is_initialized():
                ray.init(
                    address=address,
                    runtime_env={
                        "working_dir": ".",
                    },
                    ignore_reinit_error=True,
                )
            try:
                serve.start(
                    detached=True, http_options={"host": "0.0.0.0", "port": 8000}
                )
            except Exception:
                pass

            RayDeployer._initialized = True

    async def deploy(
        self,
        *,
        deployment_id: uuid.UUID,
        model_id: int,
        artifact_path: str,
        endpoint: str,
        model_type: str,
    ) -> DeploymentInfo:
        deployment_cls = _MODEL_TYPE.get(model_type)
        if not deployment_cls:
            raise UnsupportedModelType(model_type)

        deployment_name = f"model_{deployment_id}"

        ray_app = deployment_cls.options(
            name=deployment_name,
            health_check_period_s=10,
            health_check_timeout_s=30,
            max_constructor_retry_count=20,
            max_queued_requests=100,
        ).bind(artifact_path)

        try:
            _ = serve.get_app_handle(deployment_name)  # noqa: F401
            status = serve.status()
            app_status = status.applications.get(deployment_name)
            route_prefix = app_status.route_prefix if app_status else endpoint

            logger.info(
                f"Deployment {deployment_name} already exists, returning existing endpoint {route_prefix}"
            )
            return DeploymentInfo(
                endpoint=route_prefix,
                deployment_id=deployment_id,
                metadata={
                    "deployment_name": deployment_name,
                    "model_id": model_id,
                    "artifact_path": artifact_path,
                    "model_type": model_type,
                    "existing": True,
                },
            )
        except RayServeException:
            logger.info(
                f"Deployment {deployment_name} does not exist, creating new one"
            )
            pass

        serve.run(
            ray_app,
            name=deployment_name,
            route_prefix=endpoint,
            blocking=False,
        )

        logger.info(f"Successfully started deployment {deployment_name} at {endpoint}")

        full_endpoint = f"{self.base_url}{endpoint}"        

        return DeploymentInfo(
            endpoint=full_endpoint,
            deployment_id=deployment_id,
            metadata={
                "deployment_name": deployment_name,
                "model_id": model_id,
                "artifact_path": artifact_path,
                "model_type": model_type,
            },
        )

    async def undeploy(self, deployment_id: uuid.UUID) -> None:
        app_name = f"model_{deployment_id}"
        try:
            serve.get_app_handle(app_name)
            serve.delete(app_name)
            logger.info(f"Successfully deleted deployment {app_name}")
        except RayServeException:
            logger.warning(f"Deployment {app_name} does not exist, skipping delete")
            pass

    async def scale(self, deployment_id: uuid.UUID, replicas: int) -> None:
        app_name = f"model_{deployment_id}"

        try:
            app = serve.get_app_handle(app_name)
            serve.run(
                app,
                name=app_name,
                blocking=False,
                _override_deployment_options={app_name: {"num_replicas": replicas}},
            )
            logger.info(f"Scaled deployment {app_name} to {replicas} replicas")
        except RayServeException as e:
            logger.error(f"Failed to scale deployment {app_name}: {e}")
            raise
