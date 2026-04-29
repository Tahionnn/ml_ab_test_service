from typing import Any

from domain.model.entities import Model, ModelStatus
from domain.model.repository import ModelRepo
from domain.model.exceptions import (
    ModelNotFound,
    ModelCannotBeDeleted,
    InvalidStatusTransition,
    InvalidEndpointChange,
)


class ModelService:
    repo: ModelRepo

    def __init__(self, repo: ModelRepo) -> None:
        self.repo = repo

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

    async def list_models(
        self, status: ModelStatus | None = None, page: int = 1, limit: int = 20
    ) -> list[Model]:
        offset = (page - 1) * limit

        models: list[Model] = await self.repo.list_all(
            status=status, limit=limit, offset=offset
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

    def _is_valid_transition(
        self, from_status: ModelStatus, to_status: ModelStatus
    ) -> bool:
        transitions = {
            ModelStatus.STAGING: {ModelStatus.PRODUCTION},
            ModelStatus.PRODUCTION: {ModelStatus.ARCHIVED},
            ModelStatus.ARCHIVED: set(),
        }

        return to_status in transitions.get(from_status, set())
