class DomainException(Exception):
    message = "Unexpected user service error"

    def __init__(self, message: str | None = None) -> None:
        if message:
            self.message = message
        super().__init__(self.message)


class UserNotFound(DomainException):
    def __init__(self, user_id: int) -> None:
        super().__init__(f"User with {user_id} not found")


class UserAlreadyExists(DomainException):
    def __init__(self, user_id: int) -> None:
        super().__init__(f"User with {user_id} already exists")


class InvalidCredentials(DomainException):
    def __init__(self) -> None:
        super().__init__("Invalid username or password")


class InvalidToken(DomainException):
    def __init__(self) -> None:
        super().__init__("Could not validate credentials")


class ForbiddenError(DomainException):
    def __init__(self, msg: str) -> None:
        super().__init__(f"Forbidden: {msg}")
