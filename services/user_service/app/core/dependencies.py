from fastapi import Depends
from fastapi.security import OAuth2PasswordBearer
from sqlalchemy.ext.asyncio import AsyncSession

import jwt

from core.database import get_session
from core.security import decode_access_token
from domain.user.entities import User
from domain.user.exceptions import InvalidToken, UserNotFound
from domain.user.service import UserService
from infrastructure.repositories.user_repository import SQLAlchemyUserRepository


oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/auth/token")


def get_user_service(session: AsyncSession = Depends(get_session)) -> UserService:
    repo = SQLAlchemyUserRepository(session)
    return UserService(repo)


async def get_current_user(
    token: str = Depends(oauth2_scheme),
    service: UserService = Depends(get_user_service),
) -> User:
    try:
        payload = decode_access_token(token)
        user_id = int(payload["sub"])
    except (jwt.InvalidTokenError, KeyError, ValueError):
        raise InvalidToken()

    try:
        return await service.get_by_id(user_id)
    except UserNotFound:
        raise InvalidToken()
