from fastapi import APIRouter, Depends, status
from fastapi.security import OAuth2PasswordRequestForm

from core.dependencies import get_user_service
from core.security import create_access_token
from domain.user.service import UserService
from schemas.user import RegisterRequest, TokenResponse, UserResponse

auth_router = APIRouter(prefix="/auth", tags=["Auth"])


@auth_router.post(
    "/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED
)
async def register(
    request: RegisterRequest,
    service: UserService = Depends(get_user_service),
) -> UserResponse:
    user = await service.register(
        email=request.email,
        username=request.username,
        password=request.password,
        role=request.role,
    )
    return UserResponse.from_domain(usr=user)  # type: ignore[arg-type]


@auth_router.post(
    "/token", response_model=TokenResponse, status_code=status.HTTP_200_OK
)
async def login(
    form_data: OAuth2PasswordRequestForm = Depends(),
    service: UserService = Depends(get_user_service),
) -> TokenResponse:
    user = await service.authenticate(form_data.username, form_data.password)
    token = create_access_token(user_id=user.id, role=user.role.value)  # type: ignore[arg-type]
    return TokenResponse(access_token=token)
