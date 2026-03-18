package central

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"cortex/internal/models"
)

type Client interface {
	SyncArticle(ctx context.Context, payload models.CentralArticleDTO) error
}

type httpClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string, timeout time.Duration) Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	return &httpClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *httpClient) SyncArticle(ctx context.Context, payload models.CentralArticleDTO) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal article: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/articles", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create sync request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "cortex-article-sync/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("perform sync request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("central sync failed with status %d", resp.StatusCode)
	}

	return nil
}
