package modelclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type HTTPModelClient struct {
	client  *http.Client
	baseURL string
}

type BulkEndpointsResponse struct {
	Endpoints  map[string]string `json:"endpoints"`
	MissingIDs []int             `json:"missing_ids"`
}

func NewHTTPModelClient(baseURL string) *HTTPModelClient {
	return &HTTPModelClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *HTTPModelClient) GetEndpointsBulk(ctx context.Context, modelIDs []string) (map[string]string, error) {
	u, err := url.Parse(fmt.Sprintf("%s/internal/models/endpoints", c.baseURL))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for _, id := range modelIDs {
		intID, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		q.Add("model_ids", strconv.Itoa(intID))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	var out BulkEndpointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Endpoints, nil
}
