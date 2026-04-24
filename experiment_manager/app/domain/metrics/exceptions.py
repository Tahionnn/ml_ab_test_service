class DomainException(Exception):
    message = "Unexpected error has occured with metric"

    def __init__(self, message: str | None = None):
        if message:
            self.message = message
        super().__init__(self.message)


class MetricNotFound(DomainException):
    def __init__(self, metric_id: int) -> None:
        super().__init__(f"Metric with id {metric_id} not found")
