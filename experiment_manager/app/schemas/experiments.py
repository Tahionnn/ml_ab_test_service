from datetime import datetime
from pydantic import BaseModel, ConfigDict

from domain.experiments.entities import Experiment, ExperimentStatus


class CreateExperimentRequest(BaseModel):
    name: str
    description: str | None = ""
    traffic_percent: int = 100
    start_date: datetime
    end_date: datetime | None = None

    def to_domain(self) -> Experiment:
        return Experiment(
            id=None,
            name=self.name,
            description=self.description or "",
            status=ExperimentStatus.DRAFT,
            traffic_percent=self.traffic_percent,
            start_date=self.start_date,
            end_date=self.end_date,
            variants=[],
            metrics=[],
        )


class ExperimentResponse(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: int
    name: str
    description: str
    status: str
    traffic_percent: int
    start_date: datetime
    end_date: datetime | None

    @classmethod
    def from_domain(cls, exp: Experiment) -> "ExperimentResponse":
        return cls(
            id=exp.id,
            name=exp.name,
            description=exp.description,
            status=exp.status.value,
            traffic_percent=exp.traffic_percent,
            start_date=exp.start_date,
            end_date=exp.end_date,
        )


class UpdateExperimentRequest(BaseModel):
    name: str | None = None
    description: str | None = None
    traffic_percent: int | None = None
    status: ExperimentStatus | None = None
    end_date: datetime | None = None

    def merge_to_domain(self, experiment: Experiment) -> Experiment:
        if self.name is not None:
            experiment.name = self.name
        if self.description is not None:
            experiment.description = self.description
        if self.traffic_percent is not None:
            experiment.traffic_percent = self.traffic_percent
        if self.status is not None:
            experiment.status = self.status
        if self.end_date is not None:
            experiment.end_date = self.end_date
        return experiment
