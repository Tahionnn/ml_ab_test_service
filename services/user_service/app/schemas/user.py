from pydantic import BaseModel, EmailStr, Field, ConfigDict
from domain.user.entities import User, UserRole


class RegisterRequest(BaseModel):
    email: EmailStr
    username: str = Field(..., min_length=3, max_length=50)
    password: str = Field(..., min_length=8, max_length=72)
    role: UserRole = UserRole.VIEWER


class UserRoleUpdateRequest(BaseModel):
    role: UserRole


class UserResponse(BaseModel):
    id: int
    email: EmailStr
    username: str
    role: UserRole

    model_config = ConfigDict(from_attributes=True)

    @classmethod
    def from_domain(cls, usr: User) -> "UserResponse":
        return cls(id=usr.id, email=usr.email, username=usr.username, role=usr.role)


class Token(BaseModel):
    access_token: str
    token_type: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"


class InternalUserResponse(BaseModel):
    id: int
    role: UserRole

    model_config = ConfigDict(from_attributes=True)
