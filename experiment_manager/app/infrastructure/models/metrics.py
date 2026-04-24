from __future__ import annotations
from typing import TYPE_CHECKING
from sqlalchemy import String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from core.base import Base

if TYPE_CHECKING:
    from infrastructure.models.experiment_metrics import DBExperimentMetrics


class DBMetrics(Base):
    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(
        String(255), unique=True, nullable=False, index=True
    )
    type: Mapped[str] = mapped_column(String(55), nullable=False)
    formula: Mapped[str] = mapped_column(Text, nullable=False)
    unit: Mapped[str] = mapped_column(String(55), nullable=False, default="count")

    experiment_links: Mapped[list[DBExperimentMetrics]] = relationship(
        back_populates="metric"
    )
