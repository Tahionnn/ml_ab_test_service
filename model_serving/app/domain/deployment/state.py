from enum import Enum


class DeploymentStatus(str, Enum):
    REQUESTED = "requested"
    DOWNLOADING = "downloading"
    DEPLOYING = "deploying"
    READY = "ready"
    FAILED = "failed"
    UNDEPLOYING = "undeploying"
    UNDEPLOYED = "undeployed"
