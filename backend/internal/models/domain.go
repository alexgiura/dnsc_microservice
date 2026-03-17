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
	ID        uuid.UUID      `json:"id"`
	Value     string         `json:"value"` // Domain or IP
	Type      string         `json:"type"`  // "Domain" or "IP"
	Whitelist bool           `json:"whitelist"`
	Records   []DomainRecord `json:"records"`
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
