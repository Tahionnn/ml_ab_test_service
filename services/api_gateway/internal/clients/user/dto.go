package user

type UserDTO struct {
	ID       int      `json:"id"`
	Email    string   `json:"email" validate:"email"`
	Username string   `json:"username"`
	Role     UserRole `json:"role" validate:"oneof=admin scientist viewer"`
}

type RegisterRequest struct {
	Email    string   `json:"email" validate:"required,email" example:"admin@example.com"`
	Username string   `json:"username" validate:"required,min=3,max=50" example:"admin"`
	Password string   `json:"password" validate:"required,min=8,max=72" example:"strongpassword123"`
	Role     UserRole `json:"role" validate:"required,oneof=admin scientist viewer" example:"admin" enums:"admin,scientist,viewer"`
}

type LoginRequest struct {
	Username string `form:"username" example:"admin"`
	Password string `form:"password" example:"strongpassword123"`
}

type UserRoleUpdateRequest struct {
	Role UserRole `json:"role" validate:"oneof=admin scientist viewer" example:"admin"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type InternalUserResponse struct {
	ID   int      `json:"id"`
	Role UserRole `json:"role" validate:"oneof=admin scientist viewer"`
}
