package models

import (
	"time"

	"github.com/google/uuid"
)

// RTIRImportError is a row in core.rtir_import_errors (failed sync for a ticket id).
// Date is the ticket timestamp (same semantics as core.domain_records.date / RTIR LastUpdated).
type RTIRImportError struct {
	ID            uuid.UUID `json:"id"`
	TicketID      string    `json:"ticket_id"`
	Source        string    `json:"source"`
	ErrorMessage  string    `json:"error"`
	Date          time.Time `json:"date"`
	LastSyncTryAt time.Time `json:"last_sync_try_at"`
}
