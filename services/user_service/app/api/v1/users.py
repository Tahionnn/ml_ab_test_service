from fastapi import APIRouter, Depends, status

from core.dependencies import get_user_service, get_current_user
from domain.user.entities import User
from domain.user.service import UserService
from schemas.user import UserResponse, UserRoleUpdateRequest


user_router = APIRouter(prefix="/user", tags=["Users"])


@user_router.get("/me", response_model=UserResponse)
async def get_me(current_user: User = Depends(get_current_user)) -> UserResponse:
    return UserResponse.from_domain(usr=current_user)


@user_router.get("/{id}", response_model=UserResponse, status_code=status.HTTP_200_OK)
async def get_user_by_id(
    id: int,
    service: UserService = Depends(get_user_service),
    _: User = Depends(get_current_user),
) -> UserResponse:
    usr = await service.get_by_id(id)
    return UserResponse.from_domain(usr)


@user_router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_user(
    user_id: int,
    service: UserService = Depends(get_user_service),
    current_user: User = Depends(get_current_user),
) -> None:
    await service.delete(user_id, current_user)


@user_router.patch(
    "/{user_id}/role", response_model=UserResponse, status_code=status.HTTP_200_OK
)
async def update_user_role(
    user_id: int,
    request: UserRoleUpdateRequest,
    service: UserService = Depends(get_user_service),
    current_user: User = Depends(get_current_user),
) -> UserResponse:
    user = await service.update_user_role(user_id, request.role, current_user)
    return UserResponse.from_domain(user)
