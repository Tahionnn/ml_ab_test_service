package experimentclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPExperimentClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPExperimentClient(baseURL string) *HTTPExperimentClient {
	return &HTTPExperimentClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *HTTPExperimentClient) GetActiveExperiments(ctx context.Context) ([]ExperimentDTO, error) {
	url := fmt.Sprintf("%s/internal/experiments/active", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var experiments []ExperimentDTO

	if err := json.NewDecoder(resp.Body).Decode(&experiments); err != nil {
		return nil, err
	}

	return experiments, nil
}
