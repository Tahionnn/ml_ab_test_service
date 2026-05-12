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


from typing import Any

from fastapi import FastAPI, HTTPException, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from pydantic import ValidationError

from domain.user.exceptions import (
    ForbiddenError,
    InvalidCredentials,
    InvalidToken,
    UserAlreadyExists,
    UserNotFound,
)


def register_exceptions(app: FastAPI) -> None:
    def build_response(
        status_code: int,
        message: str,
        request: Request[Any],
        *,
        detail: str | None = None,
    ) -> JSONResponse:
        payload: dict[str, Any] = {
            "error": message,
            "path": str(request.url),
        }
        if detail is not None:
            payload["detail"] = detail

        return JSONResponse(
            status_code=status_code,
            content=payload,
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

    # ===== FASTAPI / PYDANTIC REQUEST VALIDATION =====
    @app.exception_handler(RequestValidationError)
    async def request_validation_handler(
        request: Request, exc: RequestValidationError
    ) -> JSONResponse:
        return build_response(
            status.HTTP_422_UNPROCESSABLE_ENTITY,
            "validation error",
            request,
            detail=str(exc),
        )

    # If you also raise ValidationError manually inside app code
    @app.exception_handler(ValidationError)
    async def validation_handler(
        request: Request, exc: ValidationError
    ) -> JSONResponse:
        return build_response(
            status.HTTP_422_UNPROCESSABLE_ENTITY,
            "validation error",
            request,
            detail=str(exc),
        )

    # ===== HTTPException (THIS WAS MISSING) =====
    @app.exception_handler(HTTPException)
    async def http_exception_handler(
        request: Request, exc: HTTPException
    ) -> JSONResponse:
        return build_response(
            exc.status_code,
            str(exc.detail),
            request,
        )

    # ===== FALLBACK =====
    @app.exception_handler(Exception)
    async def fallback_handler(request: Request, exc: Exception) -> JSONResponse:
        return build_response(
            status.HTTP_500_INTERNAL_SERVER_ERROR,
            "Internal server error",
            request,
        )
