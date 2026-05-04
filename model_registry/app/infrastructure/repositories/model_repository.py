from sqlalchemy import select, update, delete, func, or_
from sqlalchemy.ext.asyncio import AsyncSession

from infrastructure.models.model import DBModels
from domain.model.entities import Model, ModelStatus, ModelFramework
from domain.model.repository import ModelRepo
from domain.model.exceptions import ModelNotFound


class SQLAlchemyModelRepository(ModelRepo):  # type: ignore[misc]
    def __init__(self, session: AsyncSession):
        self.session = session

    async def get_by_id(self, model_id: int) -> Model | None:
        stmt = select(DBModels).where(
            DBModels.id == model_id,
        )
        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def get_by_name_and_version(self, name: str, version: str) -> Model | None:
        stmt = select(DBModels).where(
            DBModels.name == name, DBModels.version == version
        )

        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def get_by_deployment_id(self, deployment_id: int) -> Model | None:
        stmt = select(DBModels).where(DBModels.deployment_id == deployment_id)

        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()
        return self._to_domain(db_obj) if db_obj else None

    async def get_by_status(self, status: ModelStatus) -> list[Model]:
        stmt = select(DBModels).where(DBModels.status == status.value)

        result = await self.session.execute(stmt)
        db_objs = result.scalars().all()
        return [self._to_domain(db_obj) for db_obj in db_objs]

    async def search(
        self, query: str, limit: int, offset: int
    ) -> tuple[list[Model], int]:
        pattern = f"%{query.lower()}%"

        base_stmt = select(DBModels).where(
            or_(
                func.lower(DBModels.name).like(pattern),
                func.lower(DBModels.version).like(pattern),
                func.lower(DBModels.framework).like(pattern),
            )
        )

        subq = base_stmt.subquery()

        count_stmt = select(func.count()).select_from(subq)

        total_result = await self.session.execute(count_stmt)
        total = total_result.scalar_one()

        stmt = base_stmt.limit(limit).offset(offset).order_by(DBModels.id.desc())

        result = await self.session.execute(stmt)
        items = [self._to_domain(obj) for obj in result.scalars().all()]

        return items, total

    async def save(self, model: Model) -> Model:
        if model.id is None:
            db_obj = DBModels(
                name=model.name,
                version=model.version,
                artifact_uri=model.artifact_uri,
                serving_endpoint=model.serving_endpoint,
                framework=model.framework.value,
                status=model.status.value,
                deployment_id=model.deployment_id,
            )
            self.session.add(db_obj)
            await self.session.flush()
            model.id = db_obj.id
            return model
        else:
            db_obj = await self.session.get(DBModels, model.id)
            if not db_obj:
                raise ModelNotFound(model.id)

            db_obj.name = model.name
            db_obj.version = model.version
            db_obj.artifact_uri = model.artifact_uri
            db_obj.serving_endpoint = model.serving_endpoint
            db_obj.framework = model.framework.value
            db_obj.status = model.status.value
            db_obj.deployment_id = model.deployment_id

            await self.session.flush()
            return model

    async def update_status(self, model_id: int, status: ModelStatus) -> Model:
        stmt = (
            update(DBModels)
            .where(DBModels.id == model_id)
            .values(status=status.value)
            .returning(DBModels)
        )

        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()

        if db_obj is None:
            raise ModelNotFound(model_id)

        await self.session.flush()

        return self._to_domain(db_obj)

    async def update_serving_endpoint(
        self,
        model_id: int,
        endpoint: str,
    ) -> Model:
        stmt = (
            update(DBModels)
            .where(DBModels.id == model_id)
            .values(serving_endpoint=endpoint)
            .returning(DBModels)
        )

        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()

        if db_obj is None:
            raise ModelNotFound(model_id)

        await self.session.flush()

        return self._to_domain(db_obj)

    async def update_deployment_id(self, model_id: int, new_id: int) -> Model:
        stmt = (
            update(DBModels)
            .where(DBModels.id == model_id)
            .values(deployment_id=new_id)
            .returning(DBModels)
        )

        result = await self.session.execute(stmt)
        db_obj = result.scalar_one_or_none()

        if db_obj is None:
            raise ModelNotFound(model_id)

        await self.session.flush()

        return self._to_domain(db_obj)

    async def list_all(
        self,
        status: ModelStatus | None = None,
        limit: int = 20,
        offset: int = 0,
    ) -> list[Model]:
        stmt = select(DBModels)

        if status:
            stmt = stmt.where(DBModels.status == status.value)

        stmt = stmt.limit(limit).offset(offset).order_by(DBModels.id.desc())

        result = await self.session.execute(stmt)
        return [self._to_domain(obj) for obj in result.scalars().all()]

    async def delete_by_id(self, model_id: int) -> None:
        stmt = delete(DBModels).where(DBModels.id == model_id)
        await self.session.execute(stmt)
        await self.session.flush()

    def _to_domain(self, db_obj: DBModels) -> Model:
        return Model(
            id=db_obj.id,
            name=db_obj.name,
            version=db_obj.version,
            artifact_uri=db_obj.artifact_uri,
            serving_endpoint=db_obj.serving_endpoint,
            framework=ModelFramework(db_obj.framework),
            status=ModelStatus(db_obj.status),
            deployment_id=db_obj.deployment_id,
        )
