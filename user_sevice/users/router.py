from fastapi import HTTPException, Depends, APIRouter, status
from typing import Annotated, Dict, Union
from users.models.models import User
from users.models.enums import UserRole
from schemas.user import UserResponse
from auth.dependencies import get_current_user

from core.dependencies import get_user_repository
from core.repository.user_repo import UserRepository


user_router = APIRouter(prefix="/user", tags=["Users methods"])


@user_router.get("/users/me/", response_model=UserResponse)  # type: ignore[misc]
async def read_users_me(
    current_user: Annotated[User, Depends(get_current_user)],
) -> User:
    return current_user


@user_router.delete("/delete/{user_id}", status_code=status.HTTP_200_OK)  # type: ignore[misc]
async def delete_user_by_id(
    current_user: Annotated[User, Depends(get_current_user)],
    user_id: int,
    user_repo: UserRepository = Depends(get_user_repository),
) -> Dict[str, Union[str, int]]:
    if current_user.role != UserRole.ADMIN and current_user.id != user_id:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Not enough permissions"
        )
    
    user = await user_repo.get_by_id(user_id)
    
    if user is None:
        raise HTTPException(status_code=404, detail="No matches found")
    
    await user_repo.delete(user_id)
    
    return {
        "message": f"User with id={user_id} was deleted successfully",
        "user_id": user_id,
    }