from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from core.database import get_session
from infrastructure.repositories.experiments_repository import (
    SQLAlchemyExperimentRepository,
)
from infrastructure.repositories.variants_repository import SQLAlchemyVariantsRepository
from infrastructure.repositories.metrics_repository import SQLAlchemyMetricsRepository
from infrastructure.repositories.experiment_metric_repository import (
    SQLAlchemyExperimentMetricRepository,
)
from domain.experiments.service import ExperimentService
from domain.metrics.service import MetricService
from domain.experiment_metric.service import ExperimentMetricService


async def get_experiment_service(
    session: AsyncSession = Depends(get_session),
) -> ExperimentService:
    repo = SQLAlchemyExperimentRepository(session)
    variant_repo = SQLAlchemyVariantsRepository(session)
    return ExperimentService(repo, variant_repo)


async def get_metric_service(
    session: AsyncSession = Depends(get_session),
) -> MetricService:
    repo = SQLAlchemyMetricsRepository(session)
    return MetricService(repo)


async def get_experiment_metric_service(
    session: AsyncSession = Depends(get_session),
) -> ExperimentMetricService:
    repo = SQLAlchemyExperimentRepository(session)
    metric_repo = SQLAlchemyMetricsRepository(session)
    link_repo = SQLAlchemyExperimentMetricRepository(session)
    return ExperimentMetricService(repo, metric_repo, link_repo)
