from fastapi import APIRouter, Depends, status

from domain.metrics.entities import Metric
from domain.metrics.service import MetricService

from schemas.metrics import CreateMetricRequest, UpdateMetricRequest, MetricResponse

from core.dependencies import get_metric_service


metric_router = APIRouter(prefix="/metrics", tags=["Metrics"])


@metric_router.post(
    "", response_model=MetricResponse, status_code=status.HTTP_201_CREATED
)
async def create_metric(
    request: CreateMetricRequest,
    service: MetricService = Depends(get_metric_service),
) -> MetricResponse:
    metric = await service.create_metric(request.to_domain())
    return MetricResponse.from_domain(metric)


@metric_router.get(
    "", response_model=list[MetricResponse], status_code=status.HTTP_200_OK
)
async def list_metrics(
    service: MetricService = Depends(get_metric_service),
) -> list[MetricResponse]:
    metrics = await service.list_all()
    return [MetricResponse.from_domain(m) for m in metrics]


@metric_router.get(
    "/{metric_id}", response_model=MetricResponse, status_code=status.HTTP_200_OK
)
async def get_metric(
    metric_id: int,
    service: MetricService = Depends(get_metric_service),
) -> MetricService:
    metric = await service.get_by_metric_by_id(metric_id)
    return MetricResponse.from_domain(metric)


@metric_router.put(
    "/{metric_id}", response_model=MetricResponse, status_code=status.HTTP_200_OK
)
async def update_metric(
    metric_id: int,
    request: UpdateMetricRequest,
    service: MetricService = Depends(get_metric_service),
) -> MetricService:
    existing_metric = await service.get_by_metric_by_id(metric_id)

    updated_metric_data = request.apply(existing_metric)
    
    updated = await service.update_metric(updated_metric_data)

    return MetricResponse.from_domain(updated)


@metric_router.delete("/{metric_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_metric(
    metric_id: int,
    service: MetricService = Depends(get_metric_service),
) -> None:
    await service.delete_metric_by_id(metric_id)
