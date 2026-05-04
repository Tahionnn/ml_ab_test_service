from fastapi import APIRouter, Depends, status

from domain.experiments.entities import ExperimentStatus
from domain.experiments.service import ExperimentService

from schemas.internal import ExperimentWithVariantsResponse

from core.dependencies import get_experiment_service


internal_router = APIRouter(prefix="/internal", tags=["Internal"])


@internal_router.get(
    "/experiments/active",
    response_model=list[ExperimentWithVariantsResponse],
    status_code=status.HTTP_200_OK,
)
async def get_active_experiments(
    service: ExperimentService = Depends(get_experiment_service),
) -> list[ExperimentWithVariantsResponse]:
    experiments = await service.list_by_status(ExperimentStatus.RUNNING)
    return [ExperimentWithVariantsResponse.from_domain(e) for e in experiments]
