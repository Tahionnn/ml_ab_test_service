from domain.metrics.entities import Metric
from domain.metrics.repository import MetricsRepo
from domain.metrics.exceptions import MetricNotFound


class MetricService:
    def __init__(self, repo: MetricsRepo) -> None:
        self.repo = repo

    async def create_metric(self, metric: Metric) -> Metric:
        metric = await self.repo.save(metric)
        return metric

    async def list_all(self) -> list[Metric]:
        return await self.repo.list_all()

    async def get_by_metric_by_id(self, metric_id: int) -> Metric | None:
        metric = await self.repo.get_by_id(metric_id)

        if metric is None:
            raise MetricNotFound(metric_id)

        return metric

    async def delete_metric_by_id(self, metric_id: int) -> None:
        metric = await self.repo.get_by_id(metric_id)

        if metric is None:
            raise MetricNotFound(metric_id)

        await self.repo.delete_by_id(metric_id)

    async def update_metric(self, updated_metric: Metric) -> Metric:
        if updated_metric.id is None:
            raise ValueError("Metric id cannot be None")

        metric = await self.repo.get_by_id(updated_metric.id)

        if metric is None:
            raise MetricNotFound(updated_metric.id)

        metric = await self.repo.save(metric)
        return metric
