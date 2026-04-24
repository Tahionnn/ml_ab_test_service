from domain.experiments.entities import Experiment, ExperimentStatus
from domain.experiments.repository import ExperimentsRepo
from domain.experiments.exceptions import (
    ExperimentNotFound,
    ExperimentCannotBeDeleted,
    InvalidStatusTransition,
    NotEnoughVariants,
    WeightsDoNotSumTo100,
)

from domain.variants.entities import Variant
from domain.variants.repository import VariantsRepo
from domain.variants.exceptions import VariantNotFound


class ExperimentService:
    def __init__(self, repo: ExperimentsRepo, variant_repo: VariantsRepo) -> None:
        self.repo = repo
        self.variant_repo = variant_repo

    async def create_experiment(self, experiment: Experiment) -> Experiment:
        experiment = await self.repo.save(experiment)
        return experiment

    async def get_experiment_by_id(self, experiment_id: int) -> Experiment:
        experiment = await self.repo.get_by_id(experiment_id)

        if experiment is None:
            raise ExperimentNotFound(experiment_id)

        return experiment

    async def list_experiments(
        self, status: ExperimentStatus | None = None, page: int = 1, limit: int = 20
    ) -> list[Experiment]:
        offset = (page - 1) * limit
        return await self.repo.list_all(status=status, limit=limit, offset=offset)

    async def delete_experiment_by_id(self, experiment_id: int) -> Experiment:
        experiment = await self.repo.get_by_id(experiment_id)

        if experiment is None:
            raise ExperimentNotFound(experiment_id)

        if experiment.status != ExperimentStatus.DRAFT:
            raise ExperimentCannotBeDeleted(experiment.status)

        await self.repo.delete_by_id(experiment_id)

    async def update_experiment(self, updated_experiment: Experiment) -> Experiment:
        if updated_experiment.id is None:
            raise ValueError("Experiment id cannot be None")

        experiment = await self.repo.get_by_id(updated_experiment.id)

        if experiment is None:
            raise ExperimentNotFound(updated_experiment.id)

        experiment = await self.repo.save(updated_experiment)
        return experiment

    async def list_by_status(self, status: ExperimentStatus) -> list[Experiment]:
        experiments = await self.repo.list_by_status(status)
        return experiments

    async def start_experiment(self, experiment_id: int) -> Experiment:
        experiment = await self.repo.get_by_id(experiment_id)
        if experiment is None:
            raise ExperimentNotFound(experiment_id)

        if not experiment.status.can_transition_to(ExperimentStatus.RUNNING):
            raise InvalidStatusTransition(experiment.status, ExperimentStatus.RUNNING)

        if len(experiment.variants) < 2:
            raise NotEnoughVariants()

        self._validate_weights(experiment.variants)

        experiment.status = ExperimentStatus.RUNNING
        return await self.repo.save(experiment)

    async def pause_experiment(self, experiment_id: int) -> Experiment:
        experiment = await self._change_status(experiment_id, ExperimentStatus.PAUSED)
        return experiment

    async def resume_experiment(self, experiment_id: int) -> Experiment:
        experiment = await self._change_status(experiment_id, ExperimentStatus.RUNNING)
        return experiment

    async def finish_experiment(self, experiment_id: int) -> Experiment:
        experiment = await self._change_status(experiment_id, ExperimentStatus.FINISHED)
        return experiment

    async def restore_experiment(self, experiment_id: int) -> Experiment:
        experiment = await self.repo.restore_by_id(experiment_id)
        return experiment

    async def add_variant(self, experiment_id: int, variant: Variant) -> Variant:
        experiment = await self.get_experiment_by_id(experiment_id)

        if experiment.status != ExperimentStatus.DRAFT:
            raise ValueError("Cannot add variant to non-draft experiment")

        variants = await self.variant_repo.get_by_experiment_id(experiment_id)

        if variant.is_control and any(v.is_control for v in variants):
            raise ValueError("Control variant already exists")

        variant.experiment_id = experiment_id
        variants.append(variant)

        self._validate_weights_partial(variants)

        return await self.variant_repo.save(variant)

    async def list_variants(self, experiment_id: int) -> list[Variant]:
        await self.get_experiment_by_id(experiment_id)
        return await self.variant_repo.get_by_experiment_id(experiment_id)

    async def get_variant(self, variant_id: int) -> Variant:
        variant = await self.variant_repo.get_by_id(variant_id)
        if not variant:
            raise VariantNotFound(variant_id)
        return variant

    async def update_variant(self, updated_variant: Variant) -> Variant:
        if updated_variant.id is None:
            raise ValueError("Variant id cannot be None")

        variant = await self.variant_repo.get_by_id(updated_variant.id)
        if variant is None:
            raise VariantNotFound(updated_variant.id)

        experiment = await self.get_experiment_by_id(variant.experiment_id)

        if experiment.status != ExperimentStatus.DRAFT:
            raise ValueError("Cannot update variant")

        existing_in_db = await self.variant_repo.get_by_id(updated_variant.id)
        if existing_in_db.name != updated_variant.name:
            raise ValueError("Name cannot be changed")

        variant.model_id = updated_variant.model_id
        variant.traffic_weight = updated_variant.traffic_weight
        variant.is_control = updated_variant.is_control

        variants = await self.variant_repo.get_by_experiment_id(variant.experiment_id)

        variants = [v if v.id != variant.id else variant for v in variants]

        if sum(v.is_control for v in variants) > 1:
            raise ValueError("Only one control variant allowed")

        self._validate_weights_partial(variants)

        return await self.variant_repo.save(variant)

    async def delete_variant(self, variant_id: int) -> None:
        variant = await self.get_variant(variant_id)

        experiment = await self.get_experiment_by_id(variant.experiment_id)

        if experiment.status != ExperimentStatus.DRAFT:
            raise ValueError("Cannot delete variant")

        await self.variant_repo.delete_by_id(variant_id)

    @staticmethod
    def _validate_weights(variants: list[Variant]) -> None:
        total = sum(v.traffic_weight for v in variants)
        if total != 100:
            raise WeightsDoNotSumTo100(total)

    @staticmethod
    def _validate_weights_partial(variants: list[Variant]) -> None:
        total = sum(v.traffic_weight for v in variants)
        if total > 100:
            raise ValueError(f"Total weight cannot exceed 100 (got {total})")

    async def _change_status(
        self, experiment_id: int, new_status: ExperimentStatus
    ) -> Experiment:
        experiment = await self.repo.get_by_id(experiment_id)
        if not experiment:
            raise ExperimentNotFound(experiment_id)

        if not experiment.status.can_transition_to(new_status):
            raise InvalidStatusTransition(experiment.status, new_status)

        experiment.status = new_status
        return await self.repo.save(experiment)
