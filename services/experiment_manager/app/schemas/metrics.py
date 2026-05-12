from pydantic import BaseModel, Field
from domain.metrics.entities import Metric, MetricType, Unit


class CreateMetricRequest(BaseModel):
    name: str
    type: MetricType
    formula: str
    unit: Unit

    def to_domain(self) -> Metric:
        return Metric(
            id=None,
            name=self.name,
            type=self.type,
            formula=self.formula,
            unit=self.unit,
        )


class UpdateMetricRequest(BaseModel):
    name: str | None = None
    type: MetricType | None = Field(None)
    formula: str | None = None
    unit: Unit | None = Field(None)

    def apply(self, metric: Metric) -> Metric:
        for key, value in self.model_dump(exclude_unset=True).items():
            setattr(metric, key, value)
        return metric


class MetricResponse(BaseModel):
    id: int
    name: str
    type: MetricType
    formula: str
    unit: Unit

    @classmethod
    def from_domain(cls, m: Metric) -> "MetricResponse":
        return cls(
            id=m.id,
            name=m.name,
            type=m.type.value if hasattr(m.type, "value") else str(m.type),
            formula=m.formula,
            unit=m.unit.value if hasattr(m.unit, "value") else str(m.unit),
        )
