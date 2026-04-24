from dataclasses import dataclass, field
from enum import Enum

from datetime import datetime

from domain.variants.entities import Variant
from domain.experiment_metric.entities import ExperimentMetric


class ExperimentStatus(Enum):
    DRAFT = "draft"
    RUNNING = "running"
    PAUSED = "paused"
    FINISHED = "finished"

    def can_transition_to(self, target: "ExperimentStatus") -> bool:
        transitions: dict[ExperimentStatus, list[ExperimentStatus]] = {
            ExperimentStatus.DRAFT: [ExperimentStatus.RUNNING],
            ExperimentStatus.RUNNING: [
                ExperimentStatus.PAUSED,
                ExperimentStatus.FINISHED,
            ],
            ExperimentStatus.PAUSED: [
                ExperimentStatus.RUNNING,
                ExperimentStatus.FINISHED,
            ],
            ExperimentStatus.FINISHED: [],
        }
        return target in transitions.get(self, [])


@dataclass
class Experiment:
    name: str = ""
    description: str = ""
    status: ExperimentStatus = ExperimentStatus.DRAFT
    traffic_percent: int = 100
    start_date: datetime = datetime.now()
    end_date: datetime | None = None

    variants: list[Variant] = field(default_factory=list)
    metrics: list[ExperimentMetric] = field(default_factory=list)
    id: int | None = None
