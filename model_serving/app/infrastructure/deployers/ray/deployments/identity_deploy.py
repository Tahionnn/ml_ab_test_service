from typing import Any
from collections.abc import Callable
from ray import serve
from starlette.requests import Request
from .base_deployment import BaseDeployment


@serve.deployment(
    health_check_period_s=10,
    health_check_timeout_s=30,
    max_constructor_retry_count=20,
)
class IdentityModel(BaseDeployment[Callable[[Any], Any]]):
    def setup(self) -> Callable[[Any], Any]:
        print(f"Loading artifact from {self.artifact_url}")
        return lambda x: x

    async def __call__(self, request: Request) -> dict[str, Any]:
        data: Any = await request.json()

        result = self.model(data)
        return {"result": result, "url": self.artifact_url}

    def check_health(self) -> None:
        if self.model is None:
            raise RuntimeError("Model is not loaded")
