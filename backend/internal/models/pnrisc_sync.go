package models

import (
	"time"

	"github.com/google/uuid"
)

// PNRISCDomainPayload is built for outbound sync to PNRISC (domain, type, date_added, blacklisted, reason).
type PNRISCDomainPayload struct {
	DomainID    uuid.UUID
	Value       string
	Type        string
	Whitelist   bool
	DateAdded   *time.Time // latest domain_status.changed_at where whitelist=false
	ReasonTags  string     // unique tags from all domain_records, comma-separated
}
