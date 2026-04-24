from dataclasses import dataclass


@dataclass
class Variant:
    experiment_id: int
    model_id: int
    name: str
    traffic_weight: int
    is_control: bool
    description: str
    id: int | None = None
