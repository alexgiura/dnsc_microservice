package models

import (
	"time"

	"github.com/google/uuid"
)

// DomainType is Domain or IP, derived from Value
const (
	DomainTypeDomain = "Domain"
	DomainTypeIP     = "IP"
)

// Domain is the main entity: Id, Value, Type, Whitelist, and a list of records.
type Domain struct {
	ID                uuid.UUID          `json:"id"`
	Value             string             `json:"value"` // Domain or IP
	Type              string             `json:"type"`  // "Domain" or "IP"
	Whitelist         bool               `json:"whitelist"`
	Records           []DomainRecord     `json:"records"`
	StatusHistory     []DomainStatus     `json:"status_history"`
	WhitelistRequests []WhitelistRequest `json:"whitelist_requests"`
}

// DomainRecord is one record linked to a domain: TicketId, Description, Tags, Date, Source.
type DomainRecord struct {
	ID          uuid.UUID `json:"id"`
	DomainID    uuid.UUID `json:"domain_id"`
	TicketID    string    `json:"ticket_id"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Date        time.Time `json:"date"`
	Source      string    `json:"source"`
}

// DomainStatus is one record linked to a domain: each time a status changes.
type DomainStatus struct {
	ID        uuid.UUID `json:"id"`
	DomainID  uuid.UUID `json:"domain_id"`
	Whitelist bool      `json:"whitelist"`
	ChangedAt time.Time `json:"changed_at"`
	ChangedBy string    `json:"changed_by"`
	Notes     string    `json:"notes"`
}

// WhitelistDomainInput is the request body for POST /api/domains/{id}/whitelist.
// It contains the target whitelist value plus metadata about who made the change.
type WhitelistDomainInput struct {
	DomainID  *uuid.UUID `json:"domainId,omitempty"`
	Whitelist *bool      `json:"whitelist"`
	ChangeBy  string     `json:"changeBy"`
	Notes     *string    `json:"notes,omitempty"`
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
// It only exposes value/type and the last date when status became blacklist (whitelist=false).
type PublicDomain struct {
	Value string    `json:"value"`
	Type  string    `json:"type"`
	Date  time.Time `json:"date"`
}

// SaveDomainInput is the request payload for creating a domain (optionally with initial records).
type SaveDomainInput struct {
	Value     string            `json:"value"`
	Whitelist bool              `json:"whitelist"`
	Records   []SaveRecordInput `json:"records,omitempty"`
}

// SaveRecordInput is one record to add (e.g. when creating or appending to a domain).
type SaveRecordInput struct {
	TicketID    string    `json:"ticket_id"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Date        time.Time `json:"date"`
	Source      string    `json:"source"`
}

// UpdateDomainInput is the request payload for partially updating a domain (only Value and Whitelist).
type UpdateDomainInput struct {
	Value     *string `json:"value,omitempty"`
	Whitelist *bool   `json:"whitelist,omitempty"`
}
