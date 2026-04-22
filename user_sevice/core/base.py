from sqlalchemy.ext.asyncio import AsyncAttrs
from sqlalchemy.orm import DeclarativeBase, declared_attr, Mapped
from sqlalchemy.orm import mapped_column
from sqlalchemy import DateTime, func
from typing import Any
from datetime import datetime, timezone


class Base(AsyncAttrs, DeclarativeBase):  # type: ignore[misc]
    __abstract__ = True

    @declared_attr.directive  # type: ignore[misc]
    def __tablename__(cls) -> str:
        return f"{cls.__name__.lower()}"

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), 
        server_default=func.now(),
        default=lambda: datetime.now(timezone.utc)
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc)
    )