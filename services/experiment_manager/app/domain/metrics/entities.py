from dataclasses import dataclass
from enum import Enum


class MetricType(Enum):
    RATIO = "ratio"
    MEAN = "mean"
    COUNT = "count"


class Unit(Enum):
    PERCENT = "%"
    MILLISECOND = "ms"
    COUNT = "count"


@dataclass
class Metric:
    name: str
    type: MetricType
    formula: str
    unit: Unit
    id: int | None = None
