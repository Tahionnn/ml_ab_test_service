from sqlalchemy import String, Enum
from sqlalchemy.orm import Mapped, mapped_column

from core.base import Base

from domain.user.entities import UserRole


class DBUser(Base):
    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(
        String(255), unique=True, nullable=False, index=True
    )
    username: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    hashed_password: Mapped[str] = mapped_column(String(255))

    role: Mapped[UserRole] = mapped_column(
        Enum(UserRole, native_enum=False),
        server_default=UserRole.VIEWER.value,
        default=UserRole.VIEWER,
    )
