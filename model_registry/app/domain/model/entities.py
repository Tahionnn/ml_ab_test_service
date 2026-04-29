from enum import Enum
from dataclasses import dataclass


class ModelStatus(Enum):
    STAGING = "staging"
    PRODUCTION = "production"
    ARCHIVED = "archived"


@dataclass
class Model:
    name: str = ""
    version: str = ""
    artifact_uri: str = ""
    serving_endpoint: str = ""
    status: ModelStatus = ModelStatus.STAGING
    id: int | None = None
