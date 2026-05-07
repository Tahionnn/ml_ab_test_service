from pydantic import BaseModel, Field, field_validator

from domain.variants.entities import Variant


class CreateVariantRequest(BaseModel):
    name: str = Field(..., alias="variant_name")
    model_id: int
    traffic_weight: int = 1
    is_control: bool = False
    description: str | None = None

    @field_validator("traffic_weight")
    def validate_weight(cls, v: int) -> int:
        if v <= 0 or v > 100:
            raise ValueError("traffic_weight must be between 1 and 100")
        return v

    def to_domain(self) -> Variant:
        return Variant(
            id=None,
            experiment_id=0,
            name=self.name,
            model_id=self.model_id,
            traffic_weight=self.traffic_weight,
            is_control=self.is_control,
            description=self.description or "",
        )


class VariantResponse(BaseModel):
    id: int
    experiment_id: int
    name: str
    model_id: int
    traffic_weight: int
    is_control: bool
    description: str

    @classmethod
    def from_domain(cls, v: Variant) -> "VariantResponse":
        return cls(
            id=v.id,
            experiment_id=v.experiment_id,
            name=v.name,
            model_id=v.model_id,
            traffic_weight=v.traffic_weight,
            is_control=v.is_control,
            description=v.description,
        )


class UpdateVariantRequest(BaseModel):
    model_id: int | None = None
    traffic_weight: int | None = None
    is_control: bool | None = None

    def apply(self, variant: Variant) -> Variant:
        if self.model_id is not None:
            variant.model_id = self.model_id

        if self.traffic_weight is not None:
            variant.traffic_weight = self.traffic_weight

        if self.is_control is not None:
            variant.is_control = self.is_control

        return variant
