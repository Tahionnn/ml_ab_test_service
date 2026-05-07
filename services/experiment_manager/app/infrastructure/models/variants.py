from __future__ import annotations
from typing import TYPE_CHECKING
from sqlalchemy import String, Text, ForeignKey, DateTime
from sqlalchemy.orm import Mapped, mapped_column, relationship

from datetime import datetime

from core.base import Base

if TYPE_CHECKING:
    from infrastructure.models.experiments import DBExperiments


class DBExperimentVariants(Base):
    id: Mapped[int] = mapped_column(primary_key=True)
    experiment_id: Mapped[int] = mapped_column(
        ForeignKey("dbexperiments.id", ondelete="CASCADE"), nullable=False
    )
    experiment: Mapped[DBExperiments] = relationship(back_populates="variants")
    model_id: Mapped[int] = mapped_column(nullable=False)
    name: Mapped[str] = mapped_column(String(255), nullable=False, index=True)
    traffic_weight: Mapped[int] = mapped_column(nullable=False)
    is_control: Mapped[bool] = mapped_column(default=False)
    description: Mapped[str | None] = mapped_column(Text, nullable=True)
    deleted_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True, default=None
    )
