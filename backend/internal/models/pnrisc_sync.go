package models

import (
	"time"

	"github.com/google/uuid"
)

// PNRISCDomainPayload is built for outbound sync to PNRISC (domain, type, date_added, blacklisted, reason).
type PNRISCDomainPayload struct {
	DomainID   uuid.UUID
	Value      string
	Type       string
	Status    string // whitelist | blacklist | pending
	DateAdded *time.Time // latest domain_status.changed_at where status=blacklist
	Reason    string     // description from latest domain_record by date (PNRISC JSON "reason")
}
