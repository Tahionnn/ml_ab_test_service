package contextkeys

type contextKey string

const (
	UserIDKey contextKey = "user_id"

	UserRoleKey contextKey = "user_role"

	AuthTokenKey contextKey = "auth_token"

	ReqId contextKey = "request_id"
)
