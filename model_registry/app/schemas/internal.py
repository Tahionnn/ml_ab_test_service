from pydantic import BaseModel, Field, field_validator


class BulkEndpointsRequest(BaseModel):
    model_ids: list[int] = Field(
        ..., min_length=1, max_length=100, description="List of model IDs"
    )

    @field_validator("model_ids")
    @classmethod
    def validate_model_ids(cls, v: list[int]) -> list[int]:
        if len(v) != len(set(v)):
            raise ValueError("Duplicate model IDs are not allowed")
        if any(model_id <= 0 for model_id in v):
            raise ValueError("Model IDs must be positive integers")
        return v


class BulkEndpointsResponse(BaseModel):
    endpoints: dict[int, str] = Field(
        ..., description="Mapping of model_id to serving endpoint"
    )
    missing_ids: list[int] = Field(
        default_factory=list, description="Model IDs that were not found"
    )
