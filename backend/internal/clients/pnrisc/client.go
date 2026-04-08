package pnrisc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config holds PNRISC API HTTP settings.
// Token is sent as Authorization: Api-Key <token> (only the secret belongs in env).
type Config struct {
	BaseURL       string
	Token         string
	SkipTLSVerify bool
}

// Client POSTs domain payloads to the PNRISC domains API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewClient returns a PNRISC client. BaseURL should include path, e.g. https://host/core/api/domains/
func NewClient(cfg Config) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.SkipTLSVerify,
		},
	}
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second, Transport: tr},
		baseURL:    strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/") + "/",
		token:      strings.TrimSpace(cfg.Token),
	}
}

// UpsertBody matches the PNRISC API request body (date_added: YYYY-MM-DD only).
type UpsertBody struct {
	Domain      string  `json:"domain"`
	Type        string  `json:"type"`
	DateAdded   *string `json:"date_added,omitempty"`
	Blacklisted bool    `json:"blacklisted"`
	Reason      string  `json:"reason"`
}

// UpsertDomain POSTs JSON to the configured domains endpoint and returns a remote id when present in the response.
func (c *Client) UpsertDomain(ctx context.Context, body UpsertBody) (remoteID string, err error) {
	if c == nil || c.httpClient == nil {
		return "", fmt.Errorf("pnrisc client is nil")
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		auth := c.token
		if !strings.HasPrefix(strings.TrimSpace(auth), "Api-Key ") {
			auth = "Api-Key " + strings.TrimSpace(c.token)
		}
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("pnrisc: status %d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	remoteID = parseRemoteID(respBody, body.Domain)
	return remoteID, nil
}

// parseRemoteID handles a JSON object or an array of objects like:
// [{"id":"...","domain":"example.com","type":"fqdn",...}, ...]
// Prefer the row whose "domain" matches the request domain; otherwise a single row or first id.
func parseRemoteID(b []byte, matchDomain string) string {
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return ""
	}
	matchDomain = strings.TrimSpace(matchDomain)
	if b[0] == '[' {
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(b, &items); err != nil || len(items) == 0 {
			return ""
		}
		for _, m := range items {
			d := jsonStringField(m, "domain")
			if matchDomain != "" && strings.EqualFold(d, matchDomain) {
				if id := idFromMap(m); id != "" {
					return id
				}
			}
		}
		if len(items) == 1 {
			return idFromMap(items[0])
		}
		for _, m := range items {
			if id := idFromMap(m); id != "" {
				return id
			}
		}
		return ""
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil || len(m) == 0 {
		return ""
	}
	return idFromMap(m)
}

func jsonStringField(m map[string]json.RawMessage, key string) string {
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

func idFromMap(m map[string]json.RawMessage) string {
	for _, key := range []string{"id", "remote_id", "uuid", "domain_id"} {
		raw, ok := m[key]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
		var n json.Number
		if err := json.Unmarshal(raw, &n); err == nil {
			return n.String()
		}
		var f float64
		if err := json.Unmarshal(raw, &f); err == nil {
			return fmt.Sprintf("%.0f", f)
		}
	}
	return ""
}
