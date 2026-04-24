from fastapi import APIRouter, Depends, HTTPException, status

from domain.metrics.service import MetricService
from domain.metrics.exceptions import MetricNotFound

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
    try:
        metric = await service.create_metric(request.to_domain())
        return MetricResponse.from_domain(metric)

    except Exception:
        raise HTTPException(status_code=500, detail="Internal error")


@metric_router.get(
    "", response_model=list[MetricResponse], status_code=status.HTTP_200_OK
)
async def list_metrics(
    service: MetricService = Depends(get_metric_service),
) -> list[MetricResponse]:
    try:
        metrics = await service.list_all()
        return [MetricResponse.from_domain(m) for m in metrics]

    except Exception:
        raise HTTPException(status_code=500, detail="Internal error")


@metric_router.get(
    "/{metric_id}", response_model=MetricResponse, status_code=status.HTTP_200_OK
)
async def get_metric(
    metric_id: int,
    service: MetricService = Depends(get_metric_service),
) -> MetricService:
    try:
        metric = await service.get_by_metric_by_id(metric_id)
        return MetricResponse.from_domain(metric)

    except MetricNotFound:
        raise HTTPException(status_code=404, detail="Metric not found")


@metric_router.put(
    "/{metric_id}", response_model=MetricResponse, status_code=status.HTTP_200_OK
)
async def update_metric(
    metric_id: int,
    request: UpdateMetricRequest,
    service: MetricService = Depends(get_metric_service),
) -> MetricService:
    try:
        metric = await service.get_by_metric_by_id(metric_id)
        metric = request.apply(metric)

        updated = await service.update_metric(metric)

        return MetricResponse.from_domain(updated)

    except MetricNotFound:
        raise HTTPException(status_code=404, detail="Metric not found")

    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@metric_router.delete("/{metric_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_metric(
    metric_id: int,
    service: MetricService = Depends(get_metric_service),
) -> None:
    try:
        await service.delete_metric_by_id(metric_id)

    except MetricNotFound:
        raise HTTPException(status_code=404, detail="Metric not found")
