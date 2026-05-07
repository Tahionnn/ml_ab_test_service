from dataclasses import dataclass
from enum import Enum


class UserRole(str, Enum):
    ADMIN = "admin"
    SCIENTIST = "scientist"
    VIEWER = "viewer"


@dataclass
class User:
    email: str
    username: str
    hashed_password: str
    role: UserRole = UserRole.VIEWER
    id: int | None = None
