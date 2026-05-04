from abc import ABC, abstractmethod
from typing import Any, Generic, TypeVar
from starlette.requests import Request


T = TypeVar("T")


class BaseDeployment(ABC, Generic[T]):
    model: T

    def __init__(self, artifact_url: str):
        self.artifact_url = artifact_url
        self.model = self.setup()

    @abstractmethod
    def setup(self) -> T:
        raise NotImplementedError

    @abstractmethod
    async def __call__(self, request: Request) -> Any:  # type: ignore[type-arg]
        raise NotImplementedError

    @abstractmethod
    def check_health(self) -> None:
        raise NotImplementedError
