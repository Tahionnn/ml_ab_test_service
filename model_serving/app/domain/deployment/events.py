from dataclasses import dataclass


@dataclass
class DeploymentRequested:
    model_id: int
    artifact_url: str
    endpoint: str
    model_type: str


@dataclass
class ArtifactDownloaded:
    model_id: int
    local_path: str


@dataclass
class DeploymentStarted:
    model_id: int


@dataclass
class DeploymentReady:
    model_id: int
    endpoint: str


@dataclass
class DeploymentFailed:
    model_id: int
    reason: str
