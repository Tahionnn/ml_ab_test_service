from __future__ import annotations

from sqlalchemy import String, Text
from sqlalchemy.orm import Mapped, mapped_column


from core.base import Base


class DBModels(Base):  # type: ignore[misc]
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False, index=True)
    version: Mapped[str] = mapped_column(String(50), nullable=False)
    artifact_uri: Mapped[str] = mapped_column(Text)
    serving_endpoint: Mapped[str] = mapped_column(String(500))
    status: Mapped[str] = mapped_column(String(55), nullable=False, default="staging")
