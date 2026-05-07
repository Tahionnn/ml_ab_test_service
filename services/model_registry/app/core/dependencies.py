from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from core.database import get_session
from infrastructure.repositories.model_repository import SQLAlchemyModelRepository
from infrastructure.clients.bus.kafka_serving_client import (
    KafkaServingClient,
)
from domain.model.service import ModelService


async def get_model_service(
    session: AsyncSession = Depends(get_session),
) -> ModelService:
    repo = SQLAlchemyModelRepository(session)
    client = KafkaServingClient()
    return ModelService(repo, client)
