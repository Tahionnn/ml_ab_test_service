from dataclasses import dataclass


@dataclass
class ExperimentMetric:
    experiment_id: int
    metric_id: int
    goal: str
