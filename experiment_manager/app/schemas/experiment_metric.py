from pydantic import BaseModel

from domain.experiment_metric.entities import ExperimentMetric


class AttachMetricRequest(BaseModel):
    metric_id: int
    goal: str


class ExperimentMetricResponse(BaseModel):
    metric_id: int
    experiment_id: int
    goal: str

    @classmethod
    def from_domain(cls, m: ExperimentMetric) -> "ExperimentMetricResponse":
        return cls(
            metric_id=m.metric_id,
            experiment_id=m.experiment_id,
            goal=m.goal,
        )
