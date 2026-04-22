from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from core.repository.user_repo import UserRepository
from users.models.models import User
from typing import Optional, Dict, Any


class SQLAlchemyUserRepository(UserRepository):
    def __init__(self, session: AsyncSession):
        super().__init__(session)

    async def get_by_id(self, id: int) -> Optional[User]:
        stmt = select(User).where(User.id == id)
        result = await self.session.execute(stmt)
        user = result.scalar_one_or_none()
        return user if user is not None else None

    async def get_by_username_or_email(self, login: str) -> Optional[User]:
        stmt = select(User).where(
            (User.username == login) | (User.email == login)
        )
        result = await self.session.execute(stmt)
        user = result.scalar_one_or_none()
        return user if user is not None else None

    async def get_by_username(self, username: str) -> Optional[User]:
        stmt = select(User).where(User.username == username)
        result = await self.session.execute(stmt)
        user = result.scalar_one_or_none()
        return user if user is not None else None

    async def create(self, user_data: Dict[str, Any]) -> User:
        user = User(**user_data)
        self.session.add(user)
        await self.session.flush()
        await self.session.commit()
        return user

    async def delete(self, user_id: int) -> None:
        user = await self.session.get(User, user_id)
        if user:
            await self.session.delete(user)
            await self.session.commit()