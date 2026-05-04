from __future__ import annotations
import uuid

from sqlalchemy import String, Text, Uuid
from sqlalchemy.orm import Mapped, mapped_column


from core.base import Base


class DBModels(Base):  # type: ignore[misc]
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False, index=True)
    version: Mapped[str] = mapped_column(String(50), nullable=False)
    artifact_uri: Mapped[str] = mapped_column(Text)
    serving_endpoint: Mapped[str | None] = mapped_column(String(500), nullable=True)
    framework: Mapped[str] = mapped_column(
        String(50), nullable=False, default="identity"
    )
    status: Mapped[str] = mapped_column(String(55), nullable=False, default="staging")
    deployment_id: Mapped[uuid.UUID | None] = mapped_column(
        Uuid, index=True, nullable=True
    )
