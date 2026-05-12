from typing import Any, cast
import uuid

from domain.model.entities import Model, ModelStatus
from domain.model.repository import ModelRepo
from domain.model.exceptions import (
    ModelNotFound,
    ModelCannotBeUpdated,
    ModelCannotBeDeleted,
    ModelCannotBeDeploy,
    ModelCannotBeUndeployed,
    InvalidStatusTransition,
    InvalidEndpointChange,
)

from domain.serving_client.entities import (
    DeploymentInfo,
)
from domain.serving_client.client import ServingClient


class ModelService:
    repo: ModelRepo
    serving_client: ServingClient

    def __init__(self, repo: ModelRepo, serving_client: ServingClient) -> None:
        self.repo = repo
        self.serving_client = serving_client

    async def create_model(self, model: Model) -> Model:
        model = await self.repo.save(model)
        return model

    async def get_model_by_id(self, model_id: int) -> Model:
        model = await self.repo.get_by_id(model_id)

        if model is None:
            raise ModelNotFound()

        return model

    async def get_model_by_name_and_version(self, name: str, version: str) -> Model:
        model = await self.repo.get_by_name_and_version(name, version)

        if model is None:
            raise ModelNotFound(f"{name} {version}")

        return model

    async def get_model_by_deployment_id(self, deployment_id: int) -> Model:
        model = await self.repo.get_by_deployment_id(deployment_id)

        if model is None:
            raise ModelNotFound(f"{deployment_id}")

        return model

    async def get_by_status(self, status: ModelStatus) -> list[Model]:
        models = await self.repo.get_by_status(status)
        return models

    async def list_models(
        self, status: ModelStatus | None = None, page: int = 1, limit: int = 20
    ) -> list[Model]:
        offset = (page - 1) * limit

        models: list[Model] = cast(
            list[Model],
            await self.repo.list_all(status=status, limit=limit, offset=offset),
        )
        return models

    async def search_models(self, query: str, page: int, limit: int) -> dict[str, Any]:
        offset = (page - 1) * limit

        items, total = await self.repo.search(query, limit, offset)

        return {
            "items": items,
            "total": total,
        }

    async def delete_model_by_id(self, model_id: int) -> Model:
        model = await self.repo.get_by_id(model_id)

        if model is None:
            raise ModelNotFound(model_id)

        if model.status == ModelStatus.PRODUCTION:
            raise ModelCannotBeDeleted(model.status)

        await self.repo.delete_by_id(model_id)

    async def update_model(self, updated_model: Model) -> Model:
        if updated_model.id is None:
            raise ValueError("Model id cannot be None")

        model = await self.repo.get_by_id(updated_model.id)

        if model is None:
            raise ModelNotFound(updated_model.id)

        if model.status in (
            ModelStatus.ARCHIVED,
            ModelStatus.PRODUCTION,
            ModelStatus.PENDING,
        ):
            raise ModelCannotBeUpdated(updated_model.id, model.status)

        if model.deployment_id is not None:
            raise ModelCannotBeUpdated(updated_model.id, model.deployment_id)

        model = await self.repo.save(updated_model)
        return model

    async def change_status(self, model_id: int, new_status: ModelStatus) -> Model:
        model = await self.repo.get_by_id(model_id)

        if model is None:
            raise ModelNotFound(model_id)

        if model.status == new_status:
            return model

        if not self._is_valid_transition(model.status, new_status):
            raise InvalidStatusTransition(model.status, new_status)

        return await self.repo.update_status(model_id, new_status)

    async def change_serving_endpoint(
        self,
        model_id: int,
        new_endpoint: str,
    ) -> Model:
        model = await self.repo.get_by_id(model_id)

        if model is None:
            raise ModelNotFound(model_id)

        if model.serving_endpoint == new_endpoint:
            return model

        if model.status == ModelStatus.ARCHIVED:
            raise InvalidEndpointChange("Cannot change endpoint for archived model")

        if model.status == ModelStatus.PRODUCTION:
            raise InvalidEndpointChange("Cannot change endpoint in production")

        return await self.repo.update_serving_endpoint(model_id, new_endpoint)

    async def deploy_model(self, model_id: int) -> DeploymentInfo:
        model = await self.repo.get_by_id(model_id)

        if model is None:
            raise ModelNotFound(model_id)

        if model.status in (
            ModelStatus.ARCHIVED,
            ModelStatus.PRODUCTION,
            ModelStatus.PENDING,
        ):
            raise ModelCannotBeDeploy(model_id, model.status)

        if model.deployment_id is not None:
            raise ModelCannotBeDeploy(model_id, model.deployment_id)

        deployment_id = uuid.uuid4()
        generated_endpoint = f"/predict/{model.name}/{model.version}"

        deployment_result = await self.serving_client.deploy(
            deployment_id,
            model_id,
            model.artifact_uri,
            generated_endpoint,
            model.framework.value,
            metadata={},
        )

        _ = await self.repo.update_status(model_id, ModelStatus.PENDING)
        _ = await self.repo.update_deployment_id(model_id, deployment_id)

        return deployment_result

    async def undeploy_model(self, model_id: int) -> None:
        model = await self.get_model_by_id(model_id)

        if model.deployment_id is None:
            raise ModelCannotBeUndeployed(model_id)

        await self.serving_client.undeploy(model.deployment_id, model_id)

        _ = await self.repo.update_status(model_id, ModelStatus.PENDING)
        _ = await self.repo.update_deployment_id(model_id, None)

    async def on_deployment_ready(self, model_id: int, serving_endpoint: str) -> None:
        _ = await self.repo.update_status(model_id, ModelStatus.PRODUCTION)
        _ = await self.repo.update_serving_endpoint(model_id, serving_endpoint)

    async def on_deployment_failed(self, model_id: int) -> None:
        _ = await self.repo.update_status(model_id, ModelStatus.FAILED)
        _ = await self.repo.update_deployment_id(model_id, None)

    async def on_undeployment_ready(self, model_id: int) -> None:
        _ = await self.repo.update_status(model_id, ModelStatus.STAGING)
        _ = await self.repo.update_deployment_id(model_id, None)

    async def get_endpoints_bulk(self, model_ids: list[int]) -> dict[int, str]:
        if not model_ids:
            return {}

        models_list = await self.repo.get_by_ids(model_ids)

        models = {m.id: m for m in models_list}

        result = {}
        for model_id in model_ids:
            model = models.get(model_id)
            if model and model.serving_endpoint:
                result[model_id] = model.serving_endpoint
            else:
                result[model_id] = ""

        return result

    def _is_valid_transition(
        self, from_status: ModelStatus, to_status: ModelStatus
    ) -> bool:
        transitions = {
            ModelStatus.STAGING: {ModelStatus.PRODUCTION},
            ModelStatus.STAGING: {ModelStatus.ARCHIVED},
            ModelStatus.PRODUCTION: {ModelStatus.ARCHIVED},
            ModelStatus.ARCHIVED: set(),
        }

        return to_status in transitions.get(from_status, set())
