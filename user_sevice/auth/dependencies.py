# auth/dependencies.py
from fastapi import Depends, HTTPException, status
from typing import Annotated, Any
import jwt

from core.repository.user_repo import UserRepository
from core.dependencies import get_user_repository
from core.config import SECRET_KEY, ALGORITHM
from auth.utils import oauth2_scheme

async def get_current_user(
    token: Annotated[str, Depends(oauth2_scheme)],
    user_repo: UserRepository = Depends(get_user_repository),
) -> Any:
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
    )

    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise credentials_exception
    except Exception:
        raise credentials_exception

    user = await user_repo.get_by_username(username)

    if user is None:
        raise credentials_exception

    return user