from typing import Any, TYPE_CHECKING
import json
import uuid
from dataclasses import asdict
from datetime import datetime

from redis.asyncio import Redis

from domain.deployment.state_repo import DeploymentStateRepository
from domain.deployment.entities import DeploymentRecord, DeploymentStatus

from core.config import settings


if TYPE_CHECKING:
    RedisType = Redis[str]
else:
    RedisType = Redis


class RedisDeploymentStateRepository(DeploymentStateRepository):  # type: ignore
    def __init__(self, redis: RedisType | None = None):
        self._redis = redis or Redis.from_url(settings.REDIS_URL, decode_responses=True)

    def _key(self, deployment_id: str) -> str:
        return f"deployment:{deployment_id}"

    def _model_index_key(self, model_id: int) -> str:
        return f"model:{model_id}:deployments"

    async def upsert(self, record: DeploymentRecord) -> None:
        payload = asdict(record)
        dep_id_str = str(record.deployment_id)

        payload["deployment_id"] = dep_id_str
        payload["status"] = record.status.value
        payload["created_at"] = record.created_at.isoformat()
        payload["updated_at"] = record.updated_at.isoformat()

        await self._redis.set(
            self._key(dep_id_str),
            json.dumps(payload),
            ex=60 * 60 * 24 * 7,
        )
        await self._redis.sadd(self._model_index_key(record.model_id), dep_id_str)

    async def get(self, deployment_id: uuid.UUID) -> DeploymentRecord | None:
        raw = await self._redis.get(self._key(str(deployment_id)))
        if not raw:
            return None
        data = json.loads(raw)
        data["deployment_id"] = uuid.UUID(data["deployment_id"])
        data["status"] = DeploymentStatus(data["status"])
        data["created_at"] = datetime.fromisoformat(data["created_at"])
        data["updated_at"] = datetime.fromisoformat(data["updated_at"])

        return DeploymentRecord(**data)

    async def update_status(
        self,
        deployment_id: uuid.UUID,
        status: DeploymentStatus,
        *,
        error: str | None = None,
        **fields: Any,
    ) -> DeploymentRecord:
        record = await self.get(deployment_id)
        if record is None:
            raise KeyError(deployment_id)
        for k, v in fields.items():
            setattr(record, k, v)
        record.status = status
        record.error = error
        record.touch()
        await self.upsert(record)
        return record

    async def create_if_not_exists(self, record: DeploymentRecord) -> bool:
        payload = asdict(record)
        dep_id_str = str(record.deployment_id)

        payload["deployment_id"] = dep_id_str
        payload["status"] = record.status.value
        payload["created_at"] = record.created_at.isoformat()
        payload["updated_at"] = record.updated_at.isoformat()

        key = self._key(dep_id_str)

        result = await self._redis.set(
            key,
            json.dumps(payload),
            nx=True,
            ex=60 * 60 * 24 * 7,
        )

        if result:
            await self._redis.sadd(self._model_index_key(record.model_id), dep_id_str)

        return bool(result)

    async def list_by_model_id(self, model_id: int) -> list[DeploymentRecord]:
        ids = await self._redis.smembers(self._model_index_key(model_id))
        records = []
        for dep_id_str in ids:
            record = await self.get(uuid.UUID(dep_id_str))
            if record:
                records.append(record)
        return sorted(records, key=lambda r: r.updated_at, reverse=True)
