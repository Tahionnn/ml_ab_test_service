from pydantic import BaseModel
from datetime import datetime

from domain.experiments.entities import Experiment, ExperimentStatus

from schemas.variants import VariantResponse


class ExperimentWithVariantsResponse(BaseModel):
    id: int
    name: str
    description: str = ""
    status: ExperimentStatus
    traffic_percent: int
    start_date: datetime | None = None
    end_date: datetime | None = None
    variants: list[VariantResponse]

    class Config:
        from_attributes = True

    @classmethod
    def from_domain(cls, experiment: Experiment) -> "ExperimentWithVariantsResponse":
        return cls(
            id=experiment.id,
            name=experiment.name,
            description=experiment.description,
            status=experiment.status,
            traffic_percent=experiment.traffic_percent,
            start_date=experiment.start_date,
            end_date=experiment.end_date,
            variants=[
                VariantResponse(
                    id=v.id,
                    experiment_id=v.experiment_id,
                    model_id=v.model_id,
                    name=v.name,
                    traffic_weight=v.traffic_weight,
                    is_control=v.is_control,
                    description=v.description,
                )
                for v in experiment.variants
            ],
        )
