from typing import Protocol
from domain.experiment_metric.entities import ExperimentMetric


class ExperimentMetricsRepo(Protocol):
    async def save(
        self, experiment_id: int, metric_id: int, goal: str
    ) -> ExperimentMetric: ...

    async def delete(self, experiment_id: int, metric_id: int) -> None: ...

    async def list_by_experiment(
        self, experiment_id: int
    ) -> list[ExperimentMetric]: ...

    async def get(
        self, experiment_id: int, metric_id: int
    ) -> ExperimentMetric | None: ...
