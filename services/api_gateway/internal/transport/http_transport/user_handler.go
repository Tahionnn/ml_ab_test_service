package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
	mw "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport/middleware"
	"go.uber.org/zap"
)

type AuthHandler struct {
	svc *gateway.Service
	lg  *zap.Logger
}

func NewAuthHandler(svc *gateway.Service, lg *zap.Logger) *AuthHandler { return &AuthHandler{svc, lg} }

func (h *AuthHandler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Use(auth)
		r.Get("/auth/me", h.Me)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth)
		r.Get("/users/{id}", h.GetUser)
		r.Delete("/users/{id}", h.DeleteUser)
		r.Patch("/users/{id}/role", h.UpdateRole)
	})
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      user.RegisterRequest  true  "Register payload"
// @Success      201   {object}  user.UserDTO
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      409   {object}  ErrorResponse  "User already exists"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.lg.Info("auth.register.enter",
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
	)

	var req user.RegisterRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("auth.register.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("auth.register.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := user.RegisterCommand{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	h.lg.Info("auth.register.calling_service",
		zap.String("username", cmd.Username),
		zap.String("email", cmd.Email),
	)

	u, err := h.svc.Register(r.Context(), cmd)
	if err != nil {
		h.lg.Error("auth.register.failed",
			zap.Error(err),
			zap.String("username", cmd.Username),
		)
		handleErr(w, err)
		return
	}

	response := user.UserDTO{
		ID:       *u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     u.Role,
	}

	h.lg.Info("auth.register.success",
		zap.String("username", u.Username),
		zap.String("email", u.Email),
	)

	writeJSON(w, http.StatusCreated, response)
}

// Login godoc
// @Summary      Login and get JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      user.LoginRequest  true  "Login credentials"
// @Success      200   {object}  user.TokenResponse
// @Failure      400   {object}  ErrorResponse  "Invalid request body"
// @Failure      401   {object}  ErrorResponse  "Invalid credentials"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.lg.Info("auth.login.enter")

	var req user.LoginRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("auth.login.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("auth.login.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.lg.Info("auth.login.calling_service",
		zap.String("username", req.Username),
	)

	token, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		h.lg.Error("auth.login.failed",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		handleErr(w, err)
		return
	}

	response := user.TokenResponse{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
	}

	h.lg.Info("auth.login.success",
		zap.String("username", req.Username),
	)

	writeJSON(w, http.StatusOK, response)
}

// Me godoc
// @Summary      Get current user profile
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  user.UserDTO
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "User not found"
// @Router       /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r)
	if !ok {
		h.lg.Warn("auth.me.missing_user_in_context")
		writeError(w, http.StatusUnauthorized, "user not in context")
		return
	}

	h.lg.Info("auth.me.enter",
		zap.Int("user_id", userID),
	)

	u, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		h.lg.Error("auth.me.failed",
			zap.Error(err),
			zap.Int("user_id", userID),
		)
		handleErr(w, err)
		return
	}

	response := user.UserDTO{
		ID:       *u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     u.Role,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetUser godoc
// @Summary      Get user by ID
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  user.UserDTO
// @Failure      400  {object}  ErrorResponse  "Invalid user ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Forbidden"
// @Failure      404  {object}  ErrorResponse  "User not found"
// @Router       /users/{id} [get]
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("auth.get_user.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.lg.Info("auth.get_user.enter", zap.Int("id", id))

	u, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		h.lg.Error("auth.get_user.failed",
			zap.Error(err),
			zap.Int("id", id),
		)
		handleErr(w, err)
		return
	}

	response := user.UserDTO{
		ID:       *u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     u.Role,
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteUser godoc
// @Summary      Delete user (admin only)
// @Tags         users
// @Security     BearerAuth
// @Param        id   path  int  true  "User ID"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse  "Invalid user ID"
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      403  {object}  ErrorResponse  "Admin role required"
// @Failure      404  {object}  ErrorResponse  "User not found"
// @Router       /users/{id} [delete]
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("auth.delete_user.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.lg.Info("auth.delete_user.enter", zap.Int("id", id))

	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		h.lg.Error("auth.delete_user.failed",
			zap.Error(err),
			zap.Int("id", id),
		)
		handleErr(w, err)
		return
	}

	h.lg.Info("auth.delete_user.success", zap.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}

// UpdateRole godoc
// @Summary      Update user role (admin only)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                         true  "User ID"
// @Param        body  body      user.UserRoleUpdateRequest  true  "New role"
// @Success      200   {object}  user.UserDTO
// @Failure      400   {object}  ErrorResponse  "Invalid ID or request body"
// @Failure      401   {object}  ErrorResponse  "Unauthorized"
// @Failure      403   {object}  ErrorResponse  "Admin role required"
// @Failure      404   {object}  ErrorResponse  "User not found"
// @Failure      422   {object}  ErrorResponse  "Validation error"
// @Router       /users/{id}/role [patch]
func (h *AuthHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		h.lg.Warn("auth.update_role.invalid_id", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req user.UserRoleUpdateRequest
	if err := decode(r, &req); err != nil {
		h.lg.Warn("auth.update_role.decode_failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		h.lg.Warn("auth.update_role.validation_failed", zap.Error(err))
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cmd := user.UserRoleUpdateCommand{
		Role: req.Role,
	}

	h.lg.Info("auth.update_role.enter",
		zap.Int("id", id),
		zap.String("role", string(cmd.Role)),
	)

	u, err := h.svc.UpdateUserRole(r.Context(), id, cmd)
	if err != nil {
		h.lg.Error("auth.update_role.failed",
			zap.Error(err),
			zap.Int("id", id),
		)
		handleErr(w, err)
		return
	}

	response := user.UserDTO{
		ID:       *u.ID,
		Email:    u.Email,
		Username: u.Username,
		Role:     u.Role,
	}

	writeJSON(w, http.StatusOK, response)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
