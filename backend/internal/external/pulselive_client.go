package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"cortex/internal/models"
)

type PulseliveClient struct {
	baseURL    string
	httpClient *http.Client
}

type pulseliveResponse struct {
	Content  []models.ExternalArticleDTO `json:"content"`
	PageInfo struct {
		Page       int `json:"page"`
		NumPages   int `json:"numPages"`
		PageSize   int `json:"pageSize"`
		NumEntries int `json:"numEntries"`
	} `json:"pageInfo"`
}

// NewPulseliveClient creates a new client for the Pulselive content API.
func NewPulseliveClient(baseURL string, requestTimeout time.Duration) *PulseliveClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: requestTimeout,
	}

	return &PulseliveClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   requestTimeout,
			Transport: transport,
		},
	}
}

func (c *PulseliveClient) FetchLatest(ctx context.Context, page int, pageSize int) ([]models.ExternalArticleDTO, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "cortex-article-ingestor/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for GET %s", resp.StatusCode, u.String())
	}

	var payload pulseliveResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return payload.Content, nil
}

