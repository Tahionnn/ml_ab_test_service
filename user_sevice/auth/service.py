from typing import Any, Optional
from fastapi import HTTPException, status

from core.repository.user_repo import UserRepository
from auth.utils import get_password_hash, verify_password
from schemas.user import UserCreate


async def register_user_service(
    user_repo: UserRepository,
    user_data: UserCreate,
) -> Any: 
    existing_user = await user_repo.get_by_username_or_email(user_data.email)

    if existing_user:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"User with email={user_data.email} already exists",
        )

    user_dict = user_data.model_dump()
    user_dict["password"] = get_password_hash(user_data.password)

    user = await user_repo.create(user_dict)

    return user

async def authenticate_user(
    user_repo: UserRepository,
    username_or_email: str,
    password: str,
) -> Any: 
    user = await user_repo.get_by_username_or_email(username_or_email)

    if not user or not verify_password(password, user.password):
        return None

    return user