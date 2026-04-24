from pydantic import BaseModel
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
    type: str | None = None
    formula: str | None = None
    unit: str | None = None

    def apply(self, metric: Metric) -> Metric:
        if self.name is not None:
            metric.name = self.name
        if self.type is not None:
            metric.type = self.type
        if self.formula is not None:
            metric.formula = self.formula
        if self.unit is not None:
            metric.unit = self.unit
        return metric


class MetricResponse(BaseModel):
    id: int
    name: str
    type: str
    formula: str
    unit: str

    @classmethod
    def from_domain(cls, m: Metric) -> "MetricResponse":
        return cls(
            id=m.id,
            name=m.name,
            type=m.type.value if hasattr(m.type, "value") else str(m.type),
            formula=m.formula,
            unit=m.unit.value if hasattr(m.unit, "value") else str(m.unit),
        )
