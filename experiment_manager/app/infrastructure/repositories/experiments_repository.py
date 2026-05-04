from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from infrastructure.models.experiments import DBExperiments
from infrastructure.models.variants import DBExperimentVariants
from domain.experiments.repository import ExperimentsRepo
from domain.experiments.entities import Experiment, ExperimentStatus
from domain.experiments.exceptions import ExperimentNotFound
from domain.variants.entities import Variant

from datetime import datetime, timezone


class SQLAlchemyExperimentRepository(ExperimentsRepo):
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, experiment_id: int) -> Experiment | None:
        stmt = (
            select(DBExperiments)
            .where(
                DBExperiments.id == experiment_id,
                DBExperiments.deleted_at.is_(None),
            )
            .options(selectinload(DBExperiments.variants))
        )
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def save(self, experiment: Experiment) -> Experiment:
        if experiment.id is None:
            db_obj = DBExperiments(
                name=experiment.name,
                description=experiment.description,
                status=experiment.status.value,
                traffic_percent=experiment.traffic_percent,
                start_date=experiment.start_date,
                end_date=experiment.end_date,
            )
            self.session.add(db_obj)
            await self.session.flush()
            experiment.id = db_obj.id
            return experiment
        else:
            db_obj = await self.session.get(DBExperiments, experiment.id)
            if not db_obj:
                raise ExperimentNotFound(experiment.id)

            db_obj.name = experiment.name
            db_obj.description = experiment.description
            db_obj.status = experiment.status.value
            db_obj.traffic_percent = experiment.traffic_percent
            db_obj.start_date = experiment.start_date
            db_obj.end_date = experiment.end_date

            await self.session.flush()
            return experiment

    async def list_all(
        self,
        status: ExperimentStatus | None = None,
        limit: int = 20,
        offset: int = 0,
    ) -> list[Experiment]:
        stmt = (
            select(DBExperiments)
            .where(DBExperiments.deleted_at.is_(None))
            .options(selectinload(DBExperiments.variants))
        )

        if status:
            stmt = stmt.where(DBExperiments.status == status.value)

        stmt = stmt.limit(limit).offset(offset).order_by(DBExperiments.id.desc())

        result = await self.session.execute(stmt)
        return [self._to_domain(obj) for obj in result.scalars().all()]

    async def list_by_status(self, status: ExperimentStatus) -> list[Experiment]:
        stmt = (
            select(DBExperiments)
            .where(
                DBExperiments.deleted_at.is_(None), DBExperiments.status == status.value
            )
            .options(selectinload(DBExperiments.variants))
            .order_by(DBExperiments.id.desc())
        )
        result = await self.session.execute(stmt)
        return [self._to_domain(obj) for obj in result.scalars().all()]

    async def delete_by_id(self, experiment_id: int) -> None:
        stmt = (
            update(DBExperiments)
            .where(DBExperiments.id == experiment_id)
            .values(deleted_at=datetime.now(timezone.utc))
        )
        await self.session.execute(stmt)
        await self.session.flush()

    async def restore_by_id(self, experiment_id: int) -> Experiment:
        stmt = (
            select(DBExperiments)
            .where(DBExperiments.id == experiment_id)
            .options(selectinload(DBExperiments.variants))
        )
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()

        if not db_obj:
            raise ExperimentNotFound(experiment_id)

        db_obj.deleted_at = None

        await self.session.execute(
            update(DBExperimentVariants)
            .where(DBExperimentVariants.experiment_id == experiment_id)
            .values(deleted_at=None)
        )

        await self.session.flush()

        return self._to_domain(db_obj)

    def _to_domain(self, db_obj: DBExperiments) -> Experiment:
        variants = [
            Variant(
                id=v.id,
                experiment_id=v.experiment_id,
                model_id=v.model_id,
                name=v.name,
                traffic_weight=v.traffic_weight,
                is_control=v.is_control,
                description=v.description or "",
            )
            for v in (db_obj.variants or [])
            if v.deleted_at is None
        ]
        return Experiment(
            id=db_obj.id,
            name=db_obj.name,
            description=db_obj.description or "",
            status=ExperimentStatus(db_obj.status),
            traffic_percent=db_obj.traffic_percent,
            start_date=db_obj.start_date,
            end_date=db_obj.end_date,
            variants=variants,
            metrics=[],
        )
