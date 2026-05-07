class DomainException(Exception):
    message = "Unexpected error has occured with  experiment metric"

    def __init__(self, message: str | None = None):
        if message:
            self.message = message
        super().__init__(self.message)


class ExperimentMetricNotFound(DomainException):
    def __init__(self, experiment_id: int, metric_id: int) -> None:
        super().__init__(
            f"Metric {metric_id} not attached to experiment {experiment_id}"
        )


class MetricAlreadyAttached(DomainException):
    def __init__(self, experiment_id: int, metric_id: int) -> None:
        super().__init__(
            f"Metric {metric_id} already attached to experiment {experiment_id}"
        )
