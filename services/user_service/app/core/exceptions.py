from typing import Any

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from pydantic import ValidationError

from domain.user.exceptions import (
    UserNotFound,
    UserAlreadyExists,
    InvalidCredentials,
    InvalidToken,
    ForbiddenError,
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

    register(UserNotFound, 404)

    register(InvalidToken, 400)
    register(InvalidCredentials, 400)

    register(ForbiddenError, 403)

    register(UserAlreadyExists, 409)

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
