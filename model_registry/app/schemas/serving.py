from pydantic import BaseModel
import uuid


class DeployResponse(BaseModel):
    deployment_id: uuid.UUID
    status: str


class UndeployResponse(BaseModel):
    success: bool
    message: str
