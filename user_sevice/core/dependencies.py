from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from database.database import get_session
from core.repository.user_repo import UserRepository
from users.repository import SQLAlchemyUserRepository


def get_user_repository(
    session: AsyncSession = Depends(get_session),
) -> UserRepository:
    return SQLAlchemyUserRepository(session)