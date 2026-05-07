from fastapi import APIRouter, Depends

from core.security import decode_access_token
from core.dependencies import get_user_service
from domain.user.exceptions import InvalidToken
from domain.user.service import UserService
from schemas.user import InternalUserResponse

import jwt

internal_router = APIRouter(prefix="/internal", tags=["Internal"])


@internal_router.get("/verify", response_model=InternalUserResponse)
async def verify_token(
    token: str,
    service: UserService = Depends(get_user_service),
) -> InternalUserResponse:
    try:
        payload = decode_access_token(token)
        user_id = int(payload["sub"])
    except (jwt.InvalidTokenError, KeyError, ValueError):
        raise InvalidToken()

    user = await service.get_by_id(user_id)
    return InternalUserResponse(id=user.id, role=user.role)
