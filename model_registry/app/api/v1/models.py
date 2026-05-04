from fastapi import APIRouter, Depends, status, Query

from domain.model.entities import Model, ModelStatus
from domain.model.service import ModelService

from schemas.model import (
    CreateModelRequest,
    UpdateModelRequest,
    ChangeStatusRequest,
    ModelResponse,
    SearchResponse,
)
from schemas.serving import DeployResponse, UndeployResponse

from core.dependencies import get_model_service


model_router: APIRouter = APIRouter(prefix="/model")


@model_router.post(
    "", response_model=ModelResponse, status_code=status.HTTP_201_CREATED
)
async def create_model(
    request: CreateModelRequest, service: ModelService = Depends(get_model_service)
) -> ModelResponse:
    model = request.to_domain()
    created = await service.create_model(model)
    return ModelResponse.model_validate(created)


@model_router.get(
    "", response_model=list[ModelResponse], status_code=status.HTTP_200_OK
)
async def get_models(
    service: ModelService = Depends(get_model_service),
) -> ModelResponse:
    models = await service.list_models()
    return [ModelResponse.model_validate(model) for model in models]


@model_router.get(
    "/search", response_model=SearchResponse, status_code=status.HTTP_200_OK
)
async def search_models(
    q: str,
    page: int = 1,
    limit: int = 20,
    service: ModelService = Depends(get_model_service),
) -> SearchResponse:
    result = await service.search_models(q, page, limit)

    return SearchResponse(total=result["total"], models=result["items"])


@model_router.get(
    "/name/{name}/version/{version}",
    response_model=ModelResponse,
    status_code=status.HTTP_200_OK,
)
async def get_by_name_and_version(
    name: str, version: str, service: ModelService = Depends(get_model_service)
) -> ModelResponse:
    model = await service.get_model_by_name_and_version(name, version)
    return ModelResponse.model_validate(model)


@model_router.get(
    "/by-status", response_model=list[ModelResponse], status_code=status.HTTP_200_OK
)
async def get_by_status(
    status: ModelStatus = Query(..., description="Status to filter by"),
    service: ModelService = Depends(get_model_service),
) -> list[ModelResponse]:
    models = await service.get_by_status(status)
    return [ModelResponse.model_validate(model) for model in models]


@model_router.get("/{id}", response_model=ModelResponse, status_code=status.HTTP_200_OK)
async def get_by_id(
    id: int, service: ModelService = Depends(get_model_service)
) -> ModelResponse:
    model = await service.get_model_by_id(id)
    return ModelResponse.model_validate(model)


@model_router.put("/{id}", response_model=ModelResponse, status_code=status.HTTP_200_OK)
async def update_model(
    id: int,
    request: UpdateModelRequest,
    service: ModelService = Depends(get_model_service),
) -> ModelResponse:
    model = Model(id=id)
    model = request.merge_to_domain(model)

    updated = await service.update_model(model)
    return ModelResponse.model_validate(updated)


@model_router.patch(
    "/{id}/status", response_model=ModelResponse, status_code=status.HTTP_200_OK
)
async def change_status(
    id: int,
    request: ChangeStatusRequest,
    service: ModelService = Depends(get_model_service),
) -> ModelResponse:

    domain_status = request.to_domain_status()
    model = await service.change_status(id, domain_status)
    return ModelResponse.model_validate(model)


@model_router.delete("/{id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_model_by_id(
    id: int,
    service: ModelService = Depends(get_model_service),
) -> None:
    await service.delete_model_by_id(id)


@model_router.post(
    "/deploy/{model_id}", response_model=DeployResponse, status_code=status.HTTP_200_OK
)
async def deploy_model(
    id: int,
    service: ModelService = Depends(get_model_service),
) -> DeployResponse:
    info = await service.deploy_model(id)
    return DeployResponse(deployment_id=info.deployment_id, status=ModelStatus.PENDING)


@model_router.post("/undeploy/{model_id}", status_code=status.HTTP_200_OK)
async def undeploy_model(
    id: int,
    service: ModelService = Depends(get_model_service),
) -> UndeployResponse:

    await service.undeploy_model(id)

    return UndeployResponse(success=True, message="Undeployment initiated successfully")
