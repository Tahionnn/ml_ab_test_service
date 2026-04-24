from sqlalchemy import select, delete
from sqlalchemy.ext.asyncio import AsyncSession

from infrastructure.models.metrics import DBMetrics
from domain.metrics.repository import MetricsRepo
from domain.metrics.entities import Metric, MetricType, Unit


class SQLAlchemyMetricsRepository(MetricsRepo):
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, metric_id: int) -> Metric | None:
        db_obj = await self.session.get(DBMetrics, metric_id)
        return self._to_domain(db_obj) if db_obj else None

    async def list_all(self) -> list[Metric]:
        result = await self.session.execute(select(DBMetrics))
        return [self._to_domain(obj) for obj in result.scalars().all()]

    async def save(self, metric: Metric) -> Metric:
        if metric.id is None:
            db_obj = DBMetrics(
                name=metric.name,
                type=metric.type.value,
                formula=metric.formula,
                unit=metric.unit.value,
            )
            self.session.add(db_obj)
            await self.session.flush()
            metric.id = db_obj.id
            return metric
        else:
            db_obj = await self.session.get(DBMetrics, metric.id)
            if not db_obj:
                raise ValueError(f"Metric with id {metric.id} not found")

            db_obj.name = metric.name
            db_obj.type = metric.type.value
            db_obj.formula = metric.formula
            db_obj.unit = metric.unit.value

            await self.session.flush()
            return metric

    async def delete_by_id(self, metric_id: int) -> None:
        stmt = delete(DBMetrics).where(DBMetrics.id == metric_id)
        await self.session.execute(stmt)

    def _to_domain(self, db_obj: DBMetrics) -> Metric:
        return Metric(
            id=db_obj.id,
            name=db_obj.name,
            type=MetricType(db_obj.type),
            formula=db_obj.formula,
            unit=Unit(db_obj.unit),
        )
