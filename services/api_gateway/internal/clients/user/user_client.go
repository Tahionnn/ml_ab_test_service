package user

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/httpclient"
)

type UserClient interface {
	//Auth
	Register(ctx context.Context, cmd RegisterCommand) (*User, error)
	Login(ctx context.Context, username, password string) (*Token, error)

	//User
	GetUser(ctx context.Context, id int) (*User, error)
	DeleteUser(ctx context.Context, id int) error
	UpdateUserRole(ctx context.Context, id int, cmd UserRoleUpdateCommand) (*User, error)

	//Internal
	VerifyToken(ctx context.Context, token string) (*InternalUser, error)
}

type HTTPUserClient struct {
	client *httpclient.Client
}

var _ UserClient = (*HTTPUserClient)(nil)

type UserConfig struct {
	BaseURL   string
	Timeout   time.Duration
	AuthToken string
}

func NewHTTPUserClient(cfg UserConfig) *HTTPUserClient {
	return &HTTPUserClient{
		client: httpclient.New(cfg.BaseURL, cfg.Timeout, cfg.AuthToken),
	}
}

func (c *HTTPUserClient) toDomainUser(dto *UserDTO) *User {
	return &User{
		Email:    dto.Email,
		Username: dto.Username,
		Role:     dto.Role,
		ID:       &dto.ID,
	}
}

func (c *HTTPUserClient) Register(ctx context.Context, cmd RegisterCommand) (*User, error) {
	req := RegisterRequest{
		Email:    cmd.Email,
		Username: cmd.Username,
		Password: cmd.Password,
		Role:     cmd.Role,
	}

	var dto UserDTO
	err := c.client.Do(ctx, http.MethodPost, "/auth/register", req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainUser(&dto), nil
}

func (c *HTTPUserClient) Login(
	ctx context.Context,
	username,
	password string,
) (*Token, error) {

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", password)

	var dto TokenResponse

	err := c.client.DoForm(
		ctx,
		http.MethodPost,
		"/auth/token",
		form,
		&dto,
	)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}

	return &Token{
		AccessToken: dto.AccessToken,
		TokenType:   dto.TokenType,
	}, nil
}

func (c *HTTPUserClient) GetUser(ctx context.Context, id int) (*User, error) {
	var dto UserDTO
	err := c.client.Do(ctx, http.MethodGet, fmt.Sprintf("/user/%d", id), nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainUser(&dto), nil
}

func (c *HTTPUserClient) DeleteUser(ctx context.Context, id int) error {
	err := c.client.Do(ctx, http.MethodDelete, fmt.Sprintf("/user/%d", id), nil, nil)
	if err != nil {
		return c.client.MapErr(err, mapResponseError)
	}
	return nil
}
func (c *HTTPUserClient) UpdateUserRole(ctx context.Context, id int, cmd UserRoleUpdateCommand) (*User, error) {
	req := UserRoleUpdateRequest{
		Role: cmd.Role,
	}

	var dto UserDTO
	err := c.client.Do(ctx, http.MethodPatch, fmt.Sprintf("/user/%d/role", id), req, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return c.toDomainUser(&dto), nil
}

func (c *HTTPUserClient) VerifyToken(ctx context.Context, token string) (*InternalUser, error) {
	var dto InternalUserResponse
	query := url.Values{}
	query.Set("token", token)
	err := c.client.DoWithQuery(ctx, http.MethodGet, "/internal/verify", query, nil, &dto)
	if err != nil {
		return nil, c.client.MapErr(err, mapResponseError)
	}
	return &InternalUser{
		ID:   dto.ID,
		Role: dto.Role,
	}, nil
}
