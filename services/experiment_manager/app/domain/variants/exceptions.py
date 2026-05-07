class DomainException(Exception):
    message = "Unexpected error has occured with variant"

    def __init__(self, message: str | None = None):
        if message:
            self.message = message
        super().__init__(self.message)


class VariantNotFound(DomainException):
    def __init__(self, variant_id: int) -> None:
        super().__init__(f"Variant with id {variant_id} not found")
