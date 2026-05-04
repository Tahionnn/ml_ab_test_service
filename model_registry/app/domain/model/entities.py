from enum import Enum
from dataclasses import dataclass
import uuid


class ModelStatus(Enum):
    STAGING = "staging"
    PENDING = "pending"
    FAILED = "failed"
    PRODUCTION = "production"
    ARCHIVED = "archived"


class ModelFramework(Enum):
    MLFLOW = "mlflow"
    IDENTITY = "identity"
    PYTORCH = "pytorch"
    ONNX = "onnx"


@dataclass
class Model:
    name: str = ""
    version: str = ""
    artifact_uri: str = ""
    framework: ModelFramework = ModelFramework.IDENTITY
    status: ModelStatus = ModelStatus.STAGING
    id: int | None = None
    serving_endpoint: str | None = None
    deployment_id: uuid.UUID | None = None
