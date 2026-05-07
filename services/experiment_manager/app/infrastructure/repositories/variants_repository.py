from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from infrastructure.models.variants import DBExperimentVariants
from domain.variants.repository import VariantsRepo
from domain.variants.entities import Variant

from datetime import datetime, timezone


class SQLAlchemyVariantsRepository(VariantsRepo):
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, variant_id: int) -> Variant | None:
        stmt = select(DBExperimentVariants).where(
            DBExperimentVariants.id == variant_id,
            DBExperimentVariants.deleted_at.is_(None),
        )
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def get_by_experiment_id(self, experiment_id: int) -> list[Variant]:
        stmt = select(DBExperimentVariants).where(
            DBExperimentVariants.experiment_id == experiment_id,
            DBExperimentVariants.deleted_at.is_(None),
        )
        result = await self.session.execute(stmt)
        return [self._to_domain(obj) for obj in result.scalars().all()]

    async def save(self, variant: Variant) -> Variant:
        if variant.id is None:
            db_obj = DBExperimentVariants(
                experiment_id=variant.experiment_id,
                model_id=variant.model_id,
                name=variant.name,
                traffic_weight=variant.traffic_weight,
                is_control=variant.is_control,
                description=variant.description,
            )
            self.session.add(db_obj)
            await self.session.flush()
            variant.id = db_obj.id
            return variant
        else:
            db_obj = await self.session.get(DBExperimentVariants, variant.id)
            if not db_obj:
                raise ValueError(f"Variant with id {variant.id} not found")

            db_obj.model_id = variant.model_id
            db_obj.traffic_weight = variant.traffic_weight
            db_obj.is_control = variant.is_control
            db_obj.description = variant.description

            await self.session.flush()
            return variant

    async def delete_by_id(self, variant_id: int) -> None:
        stmt = (
            update(DBExperimentVariants)
            .where(
                DBExperimentVariants.id == variant_id,
                DBExperimentVariants.deleted_at.is_(None),
            )
            .values(deleted_at=datetime.now(timezone.utc))
        )
        await self.session.execute(stmt)

    def _to_domain(self, db_obj: DBExperimentVariants) -> Variant:
        return Variant(
            id=db_obj.id,
            experiment_id=db_obj.experiment_id,
            model_id=db_obj.model_id,
            name=db_obj.name,
            traffic_weight=db_obj.traffic_weight,
            is_control=db_obj.is_control,
            description=db_obj.description or "",
        )
