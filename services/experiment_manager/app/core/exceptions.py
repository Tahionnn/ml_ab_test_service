from typing import Any

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from pydantic import ValidationError

from domain.experiments.exceptions import (
    ExperimentNotFound,
    WeightsDoNotSumTo100,
    NotEnoughVariants,
    InvalidStatusTransition,
    ExperimentCannotBeDeleted,
    ExperimentAlreadyExists
)

from domain.metrics.exceptions import MetricNotFound
from domain.experiment_metric.exceptions import (
    ExperimentMetricNotFound,
    MetricAlreadyAttached,
)

from domain.variants.exceptions import VariantNotFound


def register_exceptions(app: FastAPI) -> None:

    def build_response(
        status_code: int, message: str, request: Request[Any]
    ) -> JSONResponse:
        return JSONResponse(
            status_code=status_code,
            content={
                "error": message,
                "path": str(request.url),
            },
        )

    def register(
        exc: type[Exception], status_code: int, message: str | None = None
    ) -> None:
        @app.exception_handler(exc)
        async def handler(request: Request, e: Exception) -> JSONResponse:
            return build_response(
                status_code,
                message or str(e),
                request,
            )

    # ===== DOMAIN ERRORS =====

    register(ExperimentNotFound, 404, "Experiment not found")
    register(MetricNotFound, 404, "Metric not found")
    register(VariantNotFound, 404, "Variant not found")
    register(ExperimentMetricNotFound, 404, "Metric not attached")

    register(InvalidStatusTransition, 400)
    register(ExperimentCannotBeDeleted, 400)
    register(MetricAlreadyAttached, 400)

    register(ExperimentAlreadyExists, 409)

    register(WeightsDoNotSumTo100, 422)
    register(NotEnoughVariants, 422)

    register(ValueError, 400)

    # ===== Pydantic =====
    @app.exception_handler(ValidationError)
    async def validation_handler(
        request: Request, exc: ValidationError
    ) -> JSONResponse:
        return build_response(422, str(exc), request)

    # ===== fallback =====
    @app.exception_handler(Exception)
    async def fallback_handler(request: Request, exc: Exception) -> JSONResponse:
        return build_response(500, "Internal server error", request)
