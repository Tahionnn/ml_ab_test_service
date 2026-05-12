package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/apierror"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/contextkeys"
)

type Client struct {
	baseURL string
	http    *http.Client
	token   string
}

func New(baseURL string, timeout time.Duration, token string) *Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		token:   token,
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Do(
	ctx context.Context,
	method string,
	path string,
	reqBody any,
	respBody any,
) error {
	u := c.baseURL + path

	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if authToken, ok := ctx.Value(contextkeys.AuthTokenKey).(string); ok && authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	if userID, ok := ctx.Value(contextkeys.UserIDKey).(int); ok {
		req.Header.Set("X-User-Id", strconv.Itoa(userID))
	}
	if role, ok := ctx.Value(contextkeys.UserRoleKey).(string); ok {
		req.Header.Set("X-User-Role", role)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeError(resp, path)
	}

	if respBody != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	return nil
}

func (c *Client) DoWithQuery(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	reqBody any,
	respBody any,
) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if authToken, ok := ctx.Value(contextkeys.AuthTokenKey).(string); ok && authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	if userID, ok := ctx.Value(contextkeys.UserIDKey).(int); ok {
		req.Header.Set("X-User-Id", strconv.Itoa(userID))
	}
	if role, ok := ctx.Value(contextkeys.UserRoleKey).(string); ok {
		req.Header.Set("X-User-Role", role)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeError(resp, path)
	}

	if respBody != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	return nil
}

func (c *Client) DoForm(
	ctx context.Context,
	method string,
	path string,
	form url.Values,
	respBody any,
) error {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		c.baseURL+path,
		bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if authToken, ok := ctx.Value(contextkeys.AuthTokenKey).(string); ok && authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	if userID, ok := ctx.Value(contextkeys.UserIDKey).(int); ok {
		req.Header.Set("X-User-Id", strconv.Itoa(userID))
	}
	if role, ok := ctx.Value(contextkeys.UserRoleKey).(string); ok {
		req.Header.Set("X-User-Role", role)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeError(resp, path)
	}

	if respBody != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}

	return nil
}

func (c *Client) MapErr(err error, mapResponseError func(*apierror.APIError) error) error {
	var apiErr *apierror.APIError
	if errors.As(err, &apiErr) {
		return mapResponseError(apiErr)
	}
	return err
}

func decodeError(resp *http.Response, path string) error {
	var apiErr struct {
		Error  string `json:"error"`
		Detail string `json:"detail"`
		Path   string `json:"path"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		return &apierror.APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Path:       path,
		}
	}

	msg := apiErr.Error
	if msg == "" {
		msg = apiErr.Detail
	}
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	errPath := apiErr.Path
	if errPath == "" {
		errPath = path
	}

	return &apierror.APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		Path:       errPath,
	}
}
