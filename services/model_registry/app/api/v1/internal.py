from fastapi import APIRouter, Depends, Query, status

from domain.model.service import ModelService

from schemas.internal import BulkEndpointsRequest, BulkEndpointsResponse

from core.dependencies import get_model_service


internal_router = APIRouter(prefix="/internal", tags=["Internal"])


@internal_router.get(
    "/models/endpoints",
    response_model=BulkEndpointsResponse,
    status_code=status.HTTP_200_OK,
)
async def get_endpoints_bulk(
    model_ids: list[int] = Query(..., description="Comma-separated model IDs"),
    service: ModelService = Depends(get_model_service),
) -> BulkEndpointsResponse:
    request = BulkEndpointsRequest(model_ids=model_ids)

    endpoints = await service.get_endpoints_bulk(request.model_ids)

    missing_ids = [mid for mid in request.model_ids if not endpoints.get(mid)]

    return BulkEndpointsResponse(endpoints=endpoints, missing_ids=missing_ids)
