from fastapi import APIRouter, Depends, HTTPException, Query, status

from domain.experiments.entities import Experiment, ExperimentStatus
from domain.experiments.service import ExperimentService
from domain.experiments.exceptions import (
    ExperimentNotFound,
    WeightsDoNotSumTo100,
    NotEnoughVariants,
    InvalidStatusTransition,
    ExperimentCannotBeDeleted,
)
from domain.metrics.exceptions import MetricNotFound
from domain.experiment_metric.service import ExperimentMetricService
from domain.experiment_metric.exceptions import (
    ExperimentMetricNotFound,
    MetricAlreadyAttached,
)

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

    try:
        created = await service.create_experiment(experiment)
    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except InvalidStatusTransition as e:
        raise HTTPException(status_code=400, detail=str(e))

    except (WeightsDoNotSumTo100, NotEnoughVariants) as e:
        raise HTTPException(status_code=422, detail=str(e))

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")

    return ExperimentResponse.from_domain(created)

@experiment_router.get(
    "/{id}", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def get_experriment_by_id(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    try:
        experiment = await service.get_experiment_by_id(id)
    except ExperimentNotFound:
        raise HTTPException(status_code=404)

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

    try:
        updated = await service.update_experiment(experiment)
        return ExperimentResponse.from_domain(updated)
    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except InvalidStatusTransition as e:
        raise HTTPException(status_code=400, detail=str(e))

    except (WeightsDoNotSumTo100, NotEnoughVariants) as e:
        raise HTTPException(status_code=422, detail=str(e))

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal server error. {str(e)}")


@experiment_router.delete("/{id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> None:
    try:
        await service.delete_experiment_by_id(id)
    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except ExperimentCannotBeDeleted as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@experiment_router.post(
    "/{id}/start", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def start_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    try:
        experiment = await service.start_experiment(id)
        return ExperimentResponse.from_domain(experiment)
    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except InvalidStatusTransition as e:
        raise HTTPException(status_code=400, detail=str(e))

    except (WeightsDoNotSumTo100, NotEnoughVariants) as e:
        raise HTTPException(status_code=422, detail=str(e))

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@experiment_router.post(
    "/{id}/pause", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def pause_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    try:
        experiment = await service.pause_experiment(id)
        return ExperimentResponse.from_domain(experiment)

    except ExperimentCannotBeDeleted as e:
        raise HTTPException(status_code=400, detail=str(e))

    except InvalidStatusTransition as e:
        raise HTTPException(status_code=400, detail=str(e))

    except (WeightsDoNotSumTo100, NotEnoughVariants) as e:
        raise HTTPException(status_code=422, detail=str(e))

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@experiment_router.post(
    "/{id}/resume", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def resume_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    try:
        experiment = await service.resume_experiment(id)
        return ExperimentResponse.from_domain(experiment)

    except ExperimentCannotBeDeleted as e:
        raise HTTPException(status_code=400, detail=str(e))

    except InvalidStatusTransition as e:
        raise HTTPException(status_code=400, detail=str(e))

    except (WeightsDoNotSumTo100, NotEnoughVariants) as e:
        raise HTTPException(status_code=422, detail=str(e))

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@experiment_router.post(
    "/{id}/finish", response_model=ExperimentResponse, status_code=status.HTTP_200_OK
)
async def finish_experiment(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> ExperimentResponse:
    try:
        experiment = await service.finish_experiment(id)
        return ExperimentResponse.from_domain(experiment)

    except ExperimentCannotBeDeleted as e:
        raise HTTPException(status_code=400, detail=str(e))

    except InvalidStatusTransition as e:
        raise HTTPException(status_code=400, detail=str(e))

    except (WeightsDoNotSumTo100, NotEnoughVariants) as e:
        raise HTTPException(status_code=422, detail=str(e))

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


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

    if not experiment:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Experiment not found or cannot be restored",
        )

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
    try:
        variant = request.to_domain()

        created = await service.add_variant(id, variant)

        return VariantResponse.from_domain(created)

    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@experiment_router.get(
    "/{id}/variants",
    response_model=list[VariantResponse],
    status_code=status.HTTP_200_OK,
)
async def list_variants(
    id: int,
    service: ExperimentService = Depends(get_experiment_service),
) -> list[VariantResponse]:
    try:
        variants = await service.list_variants(id)

        return [VariantResponse.from_domain(v) for v in variants]

    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


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
    try:
        result = await service.add_metric(
            experiment_id=id,
            metric_id=request.metric_id,
            goal=request.goal,
        )

        return ExperimentMetricResponse.from_domain(result)

    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except MetricNotFound:
        raise HTTPException(status_code=404, detail="Metric not found")

    except MetricAlreadyAttached:
        raise HTTPException(status_code=400, detail="Metric already attached")

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal server error {str(e)}")


@experiment_router.get(
    "/{id}/metrics",
    response_model=list[ExperimentMetricResponse],
    status_code=status.HTTP_200_OK,
)
async def list_experiment_metrics(
    id: int,
    service: ExperimentMetricService = Depends(get_experiment_metric_service),
) -> list[ExperimentMetricResponse]:
    try:
        metrics = await service.list_metrics(id)

        return [ExperimentMetricResponse.from_domain(m) for m in metrics]

    except ExperimentNotFound:
        raise HTTPException(status_code=404, detail="Experiment not found")

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")


@experiment_router.delete(
    "/{id}/metrics/{metric_id}",
    status_code=status.HTTP_204_NO_CONTENT,
)
async def remove_metric_from_experiment(
    id: int,
    metric_id: int,
    service: ExperimentMetricService = Depends(get_experiment_metric_service),
) -> None:
    try:
        await service.remove_metric(id, metric_id)

    except ExperimentMetricNotFound:
        raise HTTPException(status_code=404, detail="Metric not attached")

    except Exception:
        raise HTTPException(status_code=500, detail="Internal server error")
