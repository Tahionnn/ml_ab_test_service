from __future__ import annotations
from typing import TYPE_CHECKING

from sqlalchemy import String, ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, relationship

from core.base import Base

if TYPE_CHECKING:
    from infrastructure.models.experiments import DBExperiments
    from infrastructure.models.metrics import DBMetrics


class DBExperimentMetrics(Base):
    id: Mapped[int] = mapped_column(primary_key=True)
    experiment_id: Mapped[int] = mapped_column(
        ForeignKey("dbexperiments.id", ondelete="CASCADE"), nullable=False
    )
    metric_id: Mapped[int] = mapped_column(
        ForeignKey("dbmetrics.id", ondelete="CASCADE"), nullable=False
    )
    goal: Mapped[str] = mapped_column(
        String(55), nullable=False, default="higher_is_better"
    )

    experiment: Mapped[DBExperiments] = relationship(back_populates="metrics")
    metric: Mapped[DBMetrics] = relationship(back_populates="experiment_links")
