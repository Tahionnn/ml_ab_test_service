from fastapi import APIRouter, Depends, status

from domain.experiments.entities import Experiment
from domain.experiments.service import ExperimentService
from domain.experiment_metric.service import ExperimentMetricService

from schemas.experiments import (
    CreateExperimentRequest,
    UpdateExperimentRequest,
    ExperimentResponse,
)
from schemas.variants import CreateVariantRequest, VariantResponse
from schemas.experiment_metric import AttachMetricRequest, ExperimentMetricResponse

from core.dependencies import get_experiment_service, get_experiment_metric_service


experiment_router = APIRouter(prefix="/experiments", tags=["Experiment router"])


@experiment_router.post(
    "", response_model=ExperimentResponse, status_code=status.HTTP_201_CREATED
)
async def create_experiment(
    request: CreateExperimentRequest,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = CreateExperimentRequest.to_domain(request)
    created = await service.create_experiment(experiment)
    return ExperimentResponse.from_domain(created)


@experiment_router.get(
    "/{id}", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def get_experriment_by_id(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = await service.get_experiment_by_id(id)
    return ExperimentResponse.from_domain(experiment)


@experiment_router.put(
    "/{id}", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def update_experiment(
    id: int,
    request: UpdateExperimentRequest,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = Experiment(id=id)
    experiment = request.merge_to_domain(experiment)
    updated = await service.update_experiment(experiment)
    return ExperimentResponse.from_domain(updated)


@experiment_router.delete("/{id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> None:
    await service.delete_experiment_by_id(id)


@experiment_router.post(
    "/{id}/start", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def start_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = await service.start_experiment(id)
    return ExperimentResponse.from_domain(experiment)


@experiment_router.post(
    "/{id}/pause", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def pause_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = await service.pause_experiment(id)
    return ExperimentResponse.from_domain(experiment)


@experiment_router.post(
    "/{id}/resume", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def resume_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = await service.resume_experiment(id)
    return ExperimentResponse.from_domain(experiment)


@experiment_router.post(
    "/{id}/finish", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def finish_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = await service.finish_experiment(id)
    return ExperimentResponse.from_domain(experiment)


@experiment_router.post(
    "/{id}/restore",
    response_model=ExperimentResponse,
    status_code=status.HTTP_200_OK,
)
async def restore_experiment_endpoint(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    experiment = await service.restore_experiment(id)
    return experiment


@experiment_router.post(
    "/{id}/variants",
    response_model=VariantResponse,
    status_code=status.HTTP_201_CREATED,
)
async def add_variant(
    id: int,
    request: CreateVariantRequest,
    service: ExperimentService = Depends(get_experiment_service),
) -> VariantResponse:
    variant = request.to_domain()
    created = await service.add_variant(id, variant)
    return VariantResponse.from_domain(created)


@experiment_router.get(
    "/{id}/variants",
    response_model=list[VariantResponse],
    status_code=status.HTTP_200_OK,
)
async def list_variants(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> list[VariantResponse]:
    variants = await service.list_variants(id)
    return [VariantResponse.from_domain(v) for v in variants]


@experiment_router.post(
    "/{id}/metrics",
    response_model=ExperimentMetricResponse,
    status_code=status.HTTP_201_CREATED,
)
async def add_metric_to_experiment(
    id: int,
    request: AttachMetricRequest,
    service: ExperimentMetricService = Depends(get_experiment_metric_service),
) -> ExperimentMetricResponse:
    result = await service.add_metric(
        experiment_id=id,
        metric_id=request.metric_id,
        goal=request.goal,
    )
    return ExperimentMetricResponse.from_domain(result)


@experiment_router.get(
    "/{id}/metrics",
    response_model=list[ExperimentMetricResponse],
    status_code=status.HTTP_200_OK,
)
async def list_experiment_metrics(
    id: int,
    service: ExperimentMetricService = Depends(get_experiment_metric_service),
) -> list[ExperimentMetricResponse]:
    metrics = await service.list_metrics(id)
    return [ExperimentMetricResponse.from_domain(m) for m in metrics]


@experiment_router.delete(
    "/{id}/metrics/{metric_id}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def remove_metric_from_experiment(
    id: int,
    metric_id: int,
    service: ExperimentMetricService = Depends(get_experiment_metric_service),
) -> None:
    await service.remove_metric(id, metric_id)
