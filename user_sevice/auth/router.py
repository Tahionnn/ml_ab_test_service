from fastapi import APIRouter, Depends, HTTPException
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.ext.asyncio import AsyncSession

from typing import Annotated

from schemas.auth import Token
from schemas.user import UserCreate
from auth.service import register_user_service, authenticate_user
from auth.utils import create_access_token

from database.database import get_session

from core.dependencies import get_user_repository
from core.repository.user_repo import UserRepository


auth_router = APIRouter(tags=["Auth methods"])

@auth_router.post("/register")  # type: ignore[misc]
async def register_user(
    user_data: UserCreate,
    user_repo: UserRepository = Depends(get_user_repository),
) -> dict[str, str]:
    user = await register_user_service(user_repo, user_data)

    return {"message": f"{user.username} has been registered"}


@auth_router.post("/token")  # type: ignore[misc]
async def login(
    form_data: Annotated[OAuth2PasswordRequestForm, Depends()],
    user_repo: UserRepository = Depends(get_user_repository),
    session: AsyncSession = Depends(get_session),
) -> Token:
    user = await authenticate_user(
        user_repo,
        form_data.username,
        form_data.password,
    )

    if not user:
        raise HTTPException(status_code=401, detail="Invalid credentials")

    token = create_access_token({"sub": user.username})

    return Token(access_token=token, token_type="bearer")