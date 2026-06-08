package models

import (
	"time"

	"github.com/google/uuid"
)

// DomainType is Domain, Subdomain, or IP, derived from Value
const (
	DomainTypeDomain    = "Domain"
	DomainTypeSubdomain = "Subdomain"
	DomainTypeIP        = "IP"
)

// Domain is the main entity: value, type, status (whitelist | blacklist | pending | rejected), and records.
type Domain struct {
	ID                uuid.UUID          `json:"id"`
	Value             string             `json:"value"`                 // Domain, Subdomain, or IP
	Type              string             `json:"type"`                  // "Domain", "Subdomain", or "IP"
	Status            string             `json:"status"`                // whitelist | blacklist | pending | rejected
	Description       string             `json:"description,omitempty"` // set la primul record / primul import; nu se suprascrie la alte tichete
	Records           []DomainRecord     `json:"records"`
	StatusHistory     []DomainStatus     `json:"status_history"`
	WhitelistRequests []WhitelistRequest `json:"whitelist_requests"`
}

// DomainRecord is one record linked to a domain: TicketId, Description, Tags, Date, Source.
type DomainRecord struct {
	ID                   uuid.UUID  `json:"id"`
	DomainID             uuid.UUID  `json:"domain_id"`
	TicketID             string     `json:"ticket_id"`
	Description          string     `json:"description"`
	Tags                 []string   `json:"tags"`
	Date                 time.Time  `json:"date"`
	Source               string     `json:"source"`
	LastSuccessfulSyncAt *time.Time `json:"last_successful_sync_at,omitempty"`
}

// DomainStatus is one record linked to a domain: each time a status changes.
type DomainStatus struct {
	ID        uuid.UUID `json:"id"`
	DomainID  uuid.UUID `json:"domain_id"`
	Status    string    `json:"status"` // whitelist | blacklist | pending | rejected
	ChangedAt time.Time `json:"changed_at"`
	ChangedBy string    `json:"changed_by"`
	Notes     string    `json:"notes"`
}

// WhitelistDomainInput is the request body for POST /api/domains/{id}/whitelist.
type WhitelistDomainInput struct {
	DomainID *uuid.UUID `json:"domainId,omitempty"`
	Status   *string    `json:"status"` // whitelist | blacklist | pending | rejected
	ChangeBy string     `json:"changeBy"`
	Notes    *string    `json:"notes,omitempty"`
}

// CreateWhitelistRequestInput is the request body for submitting a whitelist request.
type CreateWhitelistRequestInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Reason    string `json:"reason"`
}

// WhitelistRequest represents one request submitted by external users.
type WhitelistRequest struct {
	ID        uuid.UUID `json:"id"`
	DomainID  uuid.UUID `json:"domain_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// PublicDomain represents the public-safe domain response.
// It only exposes value/type and the last date when status became blacklist.
type PublicDomain struct {
	Value string    `json:"value"`
	Type  string    `json:"type"`
	Date  time.Time `json:"date"`
}

// SaveDomainInput is the request payload for creating a domain (optionally with initial records).
type SaveDomainInput struct {
	Value   string            `json:"value"`
	Status  string            `json:"status,omitempty"` // whitelist | blacklist | pending | rejected; default pending
	Records []SaveRecordInput `json:"records,omitempty"`
}

// SaveRecordInput is one record to add (e.g. when creating or appending to a domain).
type SaveRecordInput struct {
	TicketID    string    `json:"ticket_id"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Date        time.Time `json:"date"`
	Source      string    `json:"source"`
}

// UpdateDomainInput is the request payload for partially updating a domain.
type UpdateDomainInput struct {
	Value       *string `json:"value,omitempty"`
	Type        *string `json:"type,omitempty"`
	Status      *string `json:"status,omitempty"`
	Description *string `json:"description,omitempty"`
}
