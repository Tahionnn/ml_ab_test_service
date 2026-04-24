from domain.experiment_metric.entities import ExperimentMetric
from domain.experiment_metric.repository import ExperimentMetricsRepo
from domain.experiment_metric.exceptions import MetricAlreadyAttached

from domain.experiments.repository import ExperimentsRepo
from domain.experiments.exceptions import ExperimentNotFound

from domain.metrics.repository import MetricsRepo
from domain.metrics.exceptions import MetricNotFound


class ExperimentMetricService:
    def __init__(
        self,
        experiment_repo: ExperimentsRepo,
        metric_repo: MetricsRepo,
        link_repo: ExperimentMetricsRepo,
    ) -> None:
        self.experiment_repo = experiment_repo
        self.metric_repo = metric_repo
        self.link_repo = link_repo

    async def add_metric(
        self,
        experiment_id: int,
        metric_id: int,
        goal: str,
    ) -> ExperimentMetric:
        experiment = await self.experiment_repo.get_by_id(experiment_id)
        if not experiment:
            raise ExperimentNotFound(experiment_id)

        metric = await self.metric_repo.get_by_id(metric_id)
        if not metric:
            raise MetricNotFound(metric_id)

        existing = await self.link_repo.get(experiment_id, metric_id)
        if existing:
            raise MetricAlreadyAttached(experiment_id, metric_id)

        return await self.link_repo.save(
            experiment_id=experiment_id,
            metric_id=metric_id,
            goal=goal,
        )

    async def remove_metric(self, experiment_id: int, metric_id: int) -> None:
        await self.link_repo.delete(experiment_id, metric_id)

    async def list_metrics(self, experiment_id: int) -> list[ExperimentMetric]:
        return await self.link_repo.list_by_experiment(experiment_id)
