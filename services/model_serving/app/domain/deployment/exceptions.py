class DomainException(Exception):
    message = "Unexpected error has occured with model"

    def __init__(self, message: str | None = None) -> None:
        if message:
            self.message = message
        super().__init__(self.message)


class UnsupportedModelType(DomainException):
    def __init__(self, model_type: str) -> None:
        super().__init__(f"Unsupported model type: {model_type}")


class DeploymentNotFound(DomainException):
    def __init__(self, deployment_id: str) -> None:
        super().__init__(f"Deployment not found: {deployment_id}")


class DeploymentAlreadyExists(RuntimeError):
    def __init__(self, deployment_name: str, existing_endpoint: str):
        self.deployment_name = deployment_name
        self.existing_endpoint = existing_endpoint
        super().__init__(
            f"Deployment '{deployment_name}' already exists at {existing_endpoint}"
        )
