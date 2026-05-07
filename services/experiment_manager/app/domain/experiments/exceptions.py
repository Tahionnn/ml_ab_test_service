from domain.experiments.entities import ExperimentStatus


class DomainException(Exception):
    message = "Unexpected error has occured with  experiment"

    def __init__(self, message: str | None = None) -> None:
        if message:
            self.message = message
        super().__init__(self.message)


class ExperimentNotFound(DomainException):
    def __init__(self, experiment_id: int) -> None:
        super().__init__(f"Experiment with id {experiment_id} not found")


class ExperimentCannotBeDeleted(DomainException):
    def __init__(self, experiment_status: ExperimentStatus) -> None:
        super().__init__(
            f"Experiment with status {experiment_status.value} cannot be deleted"
        )


class InvalidStatusTransition(DomainException):
    def __init__(self, current: ExperimentStatus, target: ExperimentStatus) -> None:
        super().__init__(f"Forbidden transition from {current.value} in {target.value}")


class WeightsDoNotSumTo100(DomainException):
    def __init__(self, actual_sum: int) -> None:
        super().__init__(f"Variant weights must sum to 100, but got {actual_sum}")


class NotEnoughVariants(DomainException):
    def __init__(self) -> None:
        super().__init__(
            "Experiment must have at least 2 variants (control + test) to start"
        )
