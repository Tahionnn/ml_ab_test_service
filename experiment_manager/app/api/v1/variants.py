from fastapi import APIRouter, Depends, status

from domain.experiments.service import ExperimentService

from schemas.variants import UpdateVariantRequest, VariantResponse

from core.dependencies import get_experiment_service


variant_router = APIRouter(prefix="/variants", tags=["Variant router"])


@variant_router.get(
    "/{variant_id}",
    response_model=VariantResponse,
    status_code=status.HTTP_200_OK,
)
async def get_variant(
    variant_id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> VariantResponse:
    variant = await service.get_variant(variant_id)
    return VariantResponse.from_domain(variant)


@variant_router.put(
    "/{variant_id}",
    response_model=VariantResponse,
    status_code=status.HTTP_200_OK,
)
async def update_variant(
    variant_id: int,
    request: UpdateVariantRequest,
    service: ExperimentService = Depends(get_experiment_service),
) -> VariantResponse:
    variant = await service.get_variant(variant_id)
    variant = request.apply(variant)
    updated = await service.update_variant(variant)
    return VariantResponse.from_domain(updated)


@variant_router.delete(
    "/{variant_id}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def delete_variant(
    variant_id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> None:
    await service.delete_variant(variant_id)
