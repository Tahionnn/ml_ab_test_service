from abc import ABC, abstractmethod
from typing import Optional, Any, Dict
from sqlalchemy.ext.asyncio import AsyncSession


class UserRepository(ABC):
    def __init__(self, session: AsyncSession):
        self.session = session

    @abstractmethod
    async def get_by_id(self, id: int) -> Optional[Any]:
        pass

    @abstractmethod
    async def get_by_username_or_email(self, login: str) -> Optional[Any]:
        pass

    @abstractmethod
    async def get_by_username(self, username: str) -> Optional[Any]:
        pass

    @abstractmethod
    async def create(self, user_data: Dict[str, Any]) -> Any:
        pass

    @abstractmethod
    async def delete(self, user_id: int) -> None:
        pass