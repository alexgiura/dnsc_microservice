package models

import (
	"time"

	"github.com/google/uuid"
)

// DashboardCounts is domain totals by status plus overall count.
type DashboardCounts struct {
	Total     int `json:"total"`
	Blacklist int `json:"blacklist"`
	Whitelist int `json:"whitelist"`
	Pending   int `json:"pending"`
	Rejected  int `json:"rejected"`
}

// DashboardRecentRecord is one of the latest core.domain_records rows with parent domain fields.
type DashboardRecentRecord struct {
	ID       uuid.UUID `json:"id"`
	DomainID uuid.UUID `json:"domain_id"`
	Value    string    `json:"value"`
	Type     string    `json:"type"`
	Status   string    `json:"status"`
	TicketID string    `json:"ticket_id"`
	Date     time.Time `json:"date"`
	Source   string    `json:"source,omitempty"`
}

// DashboardTagCount is an aggregated tag from domain_records.tags.
type DashboardTagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// DashboardBlacklistFollowUp is a domain currently in blacklist with at least one
// domain_record dated after the most recent time it was set to blacklist (in domain_status).
type DashboardBlacklistFollowUp struct {
	DomainID              uuid.UUID `json:"domain_id"`
	Value                 string    `json:"value"`
	Type                  string    `json:"type"`
	Status                string    `json:"status"`
	BlacklistedAt         time.Time `json:"blacklisted_at"`
	ReportsAfterBlacklist int       `json:"reports_after_blacklist"`
}

// DashboardResponse is GET /api/dashboard payload.
type DashboardResponse struct {
	Counts             DashboardCounts              `json:"counts"`
	RecentRecords      []DashboardRecentRecord      `json:"recent_records"`
	TopTags            []DashboardTagCount          `json:"top_tags"`
	BlacklistFollowUps []DashboardBlacklistFollowUp `json:"blacklist_follow_ups"`
}
