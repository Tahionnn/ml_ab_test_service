from fastapi import APIRouter, Depends, HTTPException, status

from domain.experiments.service import ExperimentService
from domain.experiments.exceptions import ExperimentNotFound
from domain.variants.exceptions import VariantNotFound

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
    try:
        variant = await service.get_variant(variant_id)
        return VariantResponse.from_domain(variant)

    except VariantNotFound:
        raise HTTPException(status_code=404, detail="Variant not found")


@variant_router.put(
    "/{variant_id}",
    response_model=VariantResponse,
    status_code=status.HTTP_200_OK,
)
async def update_variant(
    variant_id: int,
    request: UpdateVariantRequest,
    service: ExperimentService = Depends(get_experiment_service),
):
    try:
        variant = await service.get_variant(variant_id)

        variant = request.apply(variant)

        updated = await service.update_variant(variant)

        return VariantResponse.from_domain(updated)

    except VariantNotFound:
        raise HTTPException(status_code=404, detail="Variant not found")

    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@variant_router.delete(
    "/{variant_id}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def delete_variant(
    variant_id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> None:
    try:
        await service.delete_variant(variant_id)

    except VariantNotFound:
        raise HTTPException(status_code=404, detail="Variant not found")

    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")
