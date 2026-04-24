from __future__ import annotations
from typing import TYPE_CHECKING

from sqlalchemy import Index, String, Text, DateTime, text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from datetime import datetime

from core.base import Base

if TYPE_CHECKING:
    from infrastructure.models.variants import DBExperimentVariants
    from infrastructure.models.experiment_metrics import DBExperimentMetrics


class DBExperiments(Base):
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False, index=True)
    description: Mapped[str | None] = mapped_column(Text, nullable=True)
    status: Mapped[str] = mapped_column(String(55), nullable=False, default="draft")
    traffic_percent: Mapped[int] = mapped_column(default=100)
    start_date: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    end_date: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    deleted_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True, default=None
    )

    variants: Mapped[list[DBExperimentVariants]] = relationship(
        back_populates="experiment", cascade="all, delete-orphan"
    )
    metrics: Mapped[list[DBExperimentMetrics]] = relationship(
        back_populates="experiment", cascade="all, delete-orphan"
    )

    __table_args__ = (
        Index(
            "ix_experiments_name_active",
            "name",
            unique=True,
            postgresql_where=text("deleted_at IS NULL"),
        ),
    )
