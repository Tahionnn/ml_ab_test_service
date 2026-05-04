from enum import Enum
import re
import uuid

from pydantic import BaseModel, Field, AnyUrl, field_validator, ConfigDict


from domain.model.entities import Model, ModelFramework, ModelStatus


class ModelStatusAPI(Enum):
    STAGING = "staging"
    ARCHIVED = "archived"

    def to_domain_status(self) -> ModelStatus:
        mapping = {
            ModelStatusAPI.STAGING: ModelStatus.STAGING,
            ModelStatusAPI.ARCHIVED: ModelStatus.ARCHIVED,
        }
        return mapping[self]


class CreateModelRequest(BaseModel):
    name: str = Field(..., min_length=2, max_length=50)
    version: str
    artifact_uri: AnyUrl
    framework: ModelFramework
    status: ModelStatusAPI = Field(default=ModelStatusAPI.STAGING)

    @field_validator("version")
    @classmethod
    def validate_version(cls, v: str) -> str:
        pattern = r"^v?\d+\.\d+\.\d+$"
        if not re.match(pattern, v):
            raise ValueError("Version must match the pattern X.Y.Z or vX.Y.Z")
        return v

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "name": "image-classifier",
                "version": "v0.0.1",
                "artifact_uri": "https://models.internal/v2/res-net.zip",
                "framework": "identity",
                "status": "staging",
            }
        }
    )

    def to_domain(self) -> "Model":
        data = self.model_dump()
        data["artifact_uri"] = str(data["artifact_uri"])
        data["status"] = self.status.to_domain_status()
        return Model(**data)


class UpdateModelRequest(BaseModel):
    name: str | None = Field(
        None, min_length=2, max_length=50, examples=["new-model-name"]
    )
    version: str | None = Field(None, examples=["v1.2.4"])
    artifact_uri: AnyUrl | None = Field(
        None, examples=["https://storage.yandex/bucket/model.zip"]
    )
    framework: ModelFramework | None = Field(None, examples=[ModelFramework.IDENTITY])
    status: ModelStatusAPI | None = Field(None, examples=[ModelStatusAPI.STAGING])

    @field_validator("version")
    @classmethod
    def validate_version(cls, v: str | None) -> str | None:
        if v is not None and not re.match(r"^v?\d+\.\d+\.\d+$", v):
            raise ValueError("Version must match the pattern X.Y.Z or vX.Y.Z")
        return v

    def merge_to_domain(self, model: Model) -> Model:
        update_data = self.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            if value is not None:
                if key == "artifact_uri":
                    setattr(model, key, str(value))
                else:
                    setattr(model, key, value)

        if self.status is not None:
            model.status = self.status.to_domain_status()

        return model


class ChangeStatusRequest(BaseModel):
    status: ModelStatusAPI = Field(..., examples=[ModelStatusAPI.ARCHIVED])

    def to_domain_status(self) -> ModelStatus:
        return self.status.to_domain_status()


class ChangeEndpointRequest(BaseModel):
    serving_endpoint: str = Field(..., examples=["/api/v1/new-endpoint"])

    @field_validator("serving_endpoint")
    @classmethod
    def validate_endpoint(cls, v: str) -> str:
        if not v.startswith("/"):
            raise ValueError("serving_endpoint must start with '/'")
        return v


class ModelResponse(BaseModel):
    id: int
    name: str = Field(..., min_length=2, max_length=50)
    version: str
    artifact_uri: AnyUrl
    framework: ModelFramework
    status: ModelStatus = Field(default=ModelStatusAPI.STAGING)
    deployment_id: uuid.UUID | None

    @field_validator("version")
    @classmethod
    def validate_version(cls, v: str) -> str:
        pattern = r"^v?\d+\.\d+\.\d+$"
        if not re.match(pattern, v):
            raise ValueError("Version must match the pattern X.Y.Z or vX.Y.Z")
        return v

    model_config = ConfigDict(
        from_attributes=True,
        json_schema_extra={
            "example": {
                "id": 1,
                "name": "churn-model",
                "version": "v1.0.0",
                "artifact_uri": "s3://models/churn_v1.bin",
                "serving_endpoint": "/predict/churn",
                "framework": "identity",
                "status": "staging",
                "deployment_id": "None",
            }
        },
    )


class SearchResponse(BaseModel):
    total: int = Field(..., description="Total models found", examples=[125])
    models: list[ModelResponse]

    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "total": 1,
                "models": [
                    {
                        "id": 7,
                        "name": "recommendation-engine",
                        "version": "v2.1.0",
                        "artifact_uri": "https://ml-ops.internal/rec_v2.tar.gz",
                        "serving_endpoint": "/v2/recommend",
                        "framework": "identity",
                        "status": "production",
                    }
                ],
            }
        }
    )
