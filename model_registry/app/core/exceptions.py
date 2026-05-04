from typing import Any

from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse
from pydantic import ValidationError

from domain.model.exceptions import (
    ModelNotFound,
    ModelCannotBeUpdated,
    ModelCannotBeDeleted,
    ModelCannotBeDeploy,
    ModelCannotBeUndeployed,
    InvalidStatusTransition,
    InvalidEndpointChange,
)


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

    @app.exception_handler(ModelNotFound)
    async def model_not_found_handler(
        request: Request[Any], exc: ModelNotFound
    ) -> JSONResponse:
        return build_response(
            status.HTTP_404_NOT_FOUND,
            str(exc) or "Model not found",
            request,
        )

    @app.exception_handler(ModelCannotBeUpdated)
    async def model_cannot_be_updated_handler(
        request: Request[Any], exc: ModelCannotBeUpdated
    ) -> JSONResponse:
        return build_response(
            status.HTTP_409_CONFLICT,
            str(exc) or "Model cannot be updated",
            request,
        )

    @app.exception_handler(ModelCannotBeDeleted)
    async def model_cannot_be_deleted_handler(
        request: Request[Any], exc: ModelCannotBeDeleted
    ) -> JSONResponse:
        return build_response(
            status.HTTP_409_CONFLICT,
            str(exc) or "Model cannot be deleted",
            request,
        )

    @app.exception_handler(ModelCannotBeDeploy)
    async def model_cannot_be_deploy_handler(
        request: Request[Any], exc: ModelCannotBeDeleted
    ) -> JSONResponse:
        return build_response(
            status.HTTP_409_CONFLICT,
            str(exc) or "Model cannot be deploy",
            request,
        )

    @app.exception_handler(ModelCannotBeUndeployed)
    async def model_cannot_be_undeploy_handler(
        request: Request[Any], exc: ModelCannotBeDeleted
    ) -> JSONResponse:
        return build_response(
            status.HTTP_409_CONFLICT,
            str(exc) or "Model cannot be undeployed",
            request,
        )

    @app.exception_handler(InvalidStatusTransition)
    async def invalid_transition_handler(
        request: Request[Any], exc: InvalidStatusTransition
    ) -> JSONResponse:
        return JSONResponse(
            status_code=400,
            content={
                "error": str(exc),
            },
        )

    @app.exception_handler(InvalidEndpointChange)
    async def invalid_endpoint_handler(
        request: Request[Any], exc: InvalidEndpointChange
    ) -> JSONResponse:
        return JSONResponse(
            status_code=400,
            content={"error": str(exc)},
        )

    @app.exception_handler(ValidationError)
    async def validation_exception_handler(
        request: Request[Any], exc: ValidationError
    ) -> JSONResponse:
        return JSONResponse(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            content={
                "error": "Validation error",
                "details": exc.errors(),
            },
        )

    @app.exception_handler(Exception)
    async def global_exception_handler(
        request: Request[Any], exc: Exception
    ) -> JSONResponse:
        return build_response(
            status.HTTP_500_INTERNAL_SERVER_ERROR,
            "Internal server error",
            request,
        )
