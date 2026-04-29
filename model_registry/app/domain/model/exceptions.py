from domain.model.entities import ModelStatus


class DomainException(Exception):
    message = "Unexpected error has occured with model"

    def __init__(self, message: str | None = None) -> None:
        if message:
            self.message = message
        super().__init__(self.message)


class ModelNotFound(DomainException):
    def __init__(self, model_id: int) -> None:
        super().__init__(f"Model with {model_id} not found")


class ModelCannotBeDeleted(DomainException):
    def __init__(self, model_status: ModelStatus) -> None:
        super().__init__(f"Model with status {model_status.value} cannot be deleted")


class InvalidStatusTransition(DomainException):
    def __init__(self, from_status: ModelStatus, to_status: ModelStatus):
        self.from_status = from_status
        self.to_status = to_status

    def __str__(self) -> str:
        return f"Cannot change status from {self.from_status} to {self.to_status}"


class InvalidEndpointChange(Exception):
    def __init__(self, reason: str) -> None:
        self.reason = reason

    def __str__(self) -> str:
        return self.reason
