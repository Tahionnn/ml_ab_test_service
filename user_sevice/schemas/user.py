from pydantic import BaseModel, EmailStr, Field, ConfigDict
from typing import Optional
from users.models.enums import UserRole

class UserCreate(BaseModel): 
    email: EmailStr
    username: str
    password: str
    role: UserRole = UserRole.VIEWER

class UserResponse(BaseModel):
    id: int
    email: EmailStr
    username: str
    role: UserRole

    model_config = ConfigDict(from_attributes=True)
    