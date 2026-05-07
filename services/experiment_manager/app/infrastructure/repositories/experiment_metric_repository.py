from sqlalchemy import select, delete
from sqlalchemy.ext.asyncio import AsyncSession

from domain.experiment_metric.entities import ExperimentMetric
from domain.experiment_metric.repository import ExperimentMetricsRepo
from infrastructure.models.experiment_metrics import DBExperimentMetrics


class SQLAlchemyExperimentMetricRepository(ExperimentMetricsRepo):
    def __init__(self, session: AsyncSession):
        self.session = session

    async def save(
        self, experiment_id: int, metric_id: int, goal: str
    ) -> ExperimentMetric:

        result = await self.session.execute(
            select(DBExperimentMetrics).where(
                DBExperimentMetrics.experiment_id == experiment_id,
                DBExperimentMetrics.metric_id == metric_id,
            )
        )
        db_obj = result.scalar_one_or_none()

        if db_obj is None:
            db_obj = DBExperimentMetrics(
                experiment_id=experiment_id,
                metric_id=metric_id,
                goal=goal,
            )
            self.session.add(db_obj)
        else:
            db_obj.goal = goal

        await self.session.flush()

        return self._to_domain(db_obj)

    async def delete(self, experiment_id: int, metric_id: int) -> None:
        await self.session.execute(
            delete(DBExperimentMetrics).where(
                DBExperimentMetrics.experiment_id == experiment_id,
                DBExperimentMetrics.metric_id == metric_id,
            )
        )
        await self.session.flush()

    async def list_by_experiment(self, experiment_id: int) -> list[ExperimentMetric]:
        result = await self.session.execute(
            select(DBExperimentMetrics).where(
                DBExperimentMetrics.experiment_id == experiment_id
            )
        )
        return [self._to_domain(obj) for obj in result.scalars().all()]

    async def get(self, experiment_id: int, metric_id: int) -> ExperimentMetric | None:
        result = await self.session.execute(
            select(DBExperimentMetrics).where(
                DBExperimentMetrics.experiment_id == experiment_id,
                DBExperimentMetrics.metric_id == metric_id,
            )
        )
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    def _to_domain(self, db_obj: DBExperimentMetrics) -> ExperimentMetric:
        return ExperimentMetric(
            experiment_id=db_obj.experiment_id,
            metric_id=db_obj.metric_id,
            goal=db_obj.goal,
        )
