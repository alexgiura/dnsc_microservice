package rtir

import (
	"bytes"
	"context"
	"crypto/tls"
	"dnsc_microservice/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config holds HTTP client options for RTIR REST 2.0.
type Config struct {
	BaseURL       string
	Token         string
	SkipTLSVerify bool
	HTTPTimeout   time.Duration
}

// Client calls RTIR REST search and ticket detail endpoints.
type Client struct {
	cfg  Config
	http *http.Client
	base string
}

// NewClient builds an RTIR REST client (Authorization: token …, Accept: application/json).
func NewClient(cfg Config) *Client {
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 90 * time.Second
	}
	base := strings.TrimSpace(cfg.BaseURL)
	base = strings.TrimRight(base, "/")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipTLSVerify,
		},
	}
	return &Client{
		cfg:  cfg,
		base: base,
		http: &http.Client{Timeout: cfg.HTTPTimeout, Transport: tr},
	}
}

func (c *Client) doGET(ctx context.Context, rawURL string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "token "+strings.TrimSpace(c.cfg.Token))
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// SearchTickets returns all ticket refs matching the Updated > updatedAfter query (paginated).
func (c *Client) SearchTickets(ctx context.Context, updatedAfter time.Time, loc *time.Location) ([]models.RTIRTicketRef, error) {
	if c.base == "" {
		return nil, fmt.Errorf("rtir: empty base URL")
	}
	q := BuildSearchQuery(updatedAfter, loc)
	var all []models.RTIRTicketRef
	for page := 1; ; page++ {
		params := url.Values{}
		params.Set("query", q)
		params.Set("page", strconv.Itoa(page))
		rawURL := c.base + "/REST/2.0/tickets?" + params.Encode()

		body, status, err := c.doGET(ctx, rawURL)
		if err != nil {
			return nil, fmt.Errorf("rtir search page %d: %w", page, err)
		}
		if status < 200 || status >= 300 {
			return nil, fmt.Errorf("rtir search page %d: HTTP %d: %s", page, status, strings.TrimSpace(string(body)))
		}

		var resp models.RTIRTicketsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("rtir search decode page %d: %w", page, err)
		}
		all = append(all, resp.Items...)
		if resp.Pages == 0 {
			break
		}
		if page >= resp.Pages {
			break
		}
	}
	return all, nil
}

// GetTicketByID loads GET /REST/2.0/ticket/{id}.
func (c *Client) GetTicketByID(ctx context.Context, id string) (*models.RTIRTicketDetail, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("rtir: empty ticket id")
	}
	if c.base == "" {
		return nil, fmt.Errorf("rtir: empty base URL")
	}
	rawURL := c.base + "/REST/2.0/ticket/" + url.PathEscape(id)

	body, status, err := c.doGET(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("rtir ticket %s: %w", id, err)
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("rtir ticket %s: HTTP %d: %s", id, status, strings.TrimSpace(string(body)))
	}

	var t models.RTIRTicketDetail
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("rtir ticket %s decode: %w", id, err)
	}
	t.ID = ParseTicketID(t.IDRaw)
	return &t, nil
}

// ticketUpdateBody is the JSON body for PUT /REST/2.0/ticket/{id} to clear blacklist on the ticket.
type ticketUpdateBody struct {
	CustomFields map[string]string `json:"CustomFields"`
}

func (c *Client) UpdateTicketBlacklistNo(ctx context.Context, ticketID string) error {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return fmt.Errorf("rtir: empty ticket id")
	}
	if c.base == "" {
		return fmt.Errorf("rtir: empty base URL")
	}

	payload, err := json.Marshal(ticketUpdateBody{
		CustomFields: map[string]string{"blacklist": "no"},
	})
	if err != nil {
		return fmt.Errorf("rtir ticket %s: encode body: %w", ticketID, err)
	}

	rawURL := c.base + "/REST/2.0/ticket/" + url.PathEscape(ticketID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("rtir ticket %s: %w", ticketID, err)
	}
	req.Header.Set("Authorization", "token "+strings.TrimSpace(c.cfg.Token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("rtir ticket %s: %w", ticketID, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("rtir ticket %s: read body: %w", ticketID, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rtir ticket %s: HTTP %d: %s", ticketID, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil
	}
	var msgs []string
	if err := json.Unmarshal(body, &msgs); err != nil {
		return fmt.Errorf("rtir ticket %s: decode response %q: %w", ticketID, trimmed, err)
	}
	if len(msgs) == 0 {
		// RT often returns [] when blacklist was already "no" (e.g. cleared in the UI).
		return nil
	}
	for _, m := range msgs {
		low := strings.ToLower(m)
		if strings.Contains(low, "blacklist") && strings.Contains(low, "no") {
			return nil
		}
	}
	return fmt.Errorf("rtir ticket %s: unexpected RT response: %v", ticketID, msgs)
}
