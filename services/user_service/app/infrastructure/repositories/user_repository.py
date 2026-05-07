from sqlalchemy import select, delete
from sqlalchemy.ext.asyncio import AsyncSession

from domain.user.entities import User
from domain.user.repository import UserRepo
from domain.user.exceptions import UserNotFound

from infrastructure.models.user import DBUser


class SQLAlchemyUserRepository(UserRepo):  # type: ignore[misc]
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, id: int) -> User | None:
        stmt = select(DBUser).where(DBUser.id == id)
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def get_by_username_or_email(self, login: str) -> User | None:
        stmt = select(DBUser).where(
            (DBUser.username == login) | (DBUser.email == login)
        )
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def get_by_username(self, username: str) -> User | None:
        stmt = select(DBUser).where(DBUser.username == username)
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def save(self, user: User) -> User:
        if user.id is None:
            db_obj = DBUser(
                email=user.email,
                username=user.username,
                hashed_password=user.hashed_password,
                role=user.role,
            )
            self.session.add(db_obj)
            await self.session.flush()
            user.id = db_obj.id
            return user
        else:
            db_obj = await self.session.get(DBUser, user.id)
            if not db_obj:
                raise UserNotFound(user.id)

            db_obj.email = user.email
            db_obj.username = user.username
            db_obj.hashed_password = user.hashed_password
            db_obj.role = user.role

            await self.session.flush()
            return user

    async def delete(self, user_id: int) -> None:
        stmt = delete(DBUser).where(DBUser.id == user_id)
        await self.session.execute(stmt)
        await self.session.flush()

    def _to_domain(self, db_obj: DBUser) -> User:
        return User(
            email=db_obj.email,
            username=db_obj.username,
            hashed_password=db_obj.hashed_password,
            role=db_obj.role,
            id=db_obj.id,
        )
