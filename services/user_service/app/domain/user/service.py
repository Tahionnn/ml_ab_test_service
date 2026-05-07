from domain.user.entities import User, UserRole
from domain.user.repository import UserRepo
from domain.user.exceptions import (
    UserNotFound,
    UserAlreadyExists,
    InvalidCredentials,
    ForbiddenError,
)
from core.security import hash_password, verify_password


class UserService:
    def __init__(self, repo: UserRepo) -> None:
        self.repo = repo

    async def register(
        self, email: str, username: str, password: str, role: UserRole = UserRole.VIEWER
    ) -> User:
        existing = await self.repo.get_by_username_or_email(email)
        if existing is not None:
            raise UserAlreadyExists(email)

        user = User(
            email=email,
            username=username,
            hashed_password=hash_password(password),
            role=role,
        )
        return await self.repo.save(user)

    async def authenticate(self, login: str, password: str) -> User:
        user = await self.repo.get_by_username_or_email(login)
        if user is None or not verify_password(password, user.hashed_password):
            raise InvalidCredentials()
        return user

    async def get_by_id(self, user_id: int) -> User:
        user = await self.repo.get_by_id(user_id)
        if user is None:
            raise UserNotFound(user_id)
        return user

    async def get_by_username(self, username: str) -> User:
        user = await self.repo.get_by_username(username)
        if user is None:
            raise UserNotFound(username)
        return user

    async def delete(self, user_id: int, current_user: User) -> None:
        user_to_delete = await self.repo.get_by_id(user_id)
        if user_to_delete is None:
            raise UserNotFound(user_id)

        if current_user.role != UserRole.ADMIN:
            raise ForbiddenError("Only administrators can delete users")

        if current_user.id == user_id:
            raise ForbiddenError("Cannot delete your own account")

        await self.repo.delete(user_id)

    async def update_user_role(
        self, user_id: int, new_role: UserRole, current_user: User
    ) -> User:
        if current_user.role != UserRole.ADMIN:
            raise ForbiddenError("Only administrators can change user roles")

        user = await self.repo.get_by_id(user_id)
        if user is None:
            raise UserNotFound(user_id)

        user.role = new_role
        return await self.repo.save(user)
