package models

import (
	"encoding/json"
	"time"
)

// RTIRPlayTicketsResponse is the RTIR REST 2.0 /tickets search JSON body.
type RTIRPlayTicketsResponse struct {
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	Count   int             `json:"count"`
	Pages   int             `json:"pages"`
	PerPage int             `json:"per_page"`
	Items   []RTIRTicketRef `json:"items"`
}

// RTIRTicketRef is a ticket stub from /REST/2.0/tickets search.
type RTIRTicketRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	URL  string `json:"_url"`
}

// RTIRCustomField is one entry in ticket CustomFields.
type RTIRCustomField struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// RTIRTicketDetail is GET /REST/2.0/ticket/{id} (id may be JSON number or string).
type RTIRTicketDetail struct {
	IDRaw        json.RawMessage    `json:"id"`
	ID           string             `json:"-"`
	Subject      string             `json:"Subject"`
	LastUpdated  string             `json:"LastUpdated"`
	CustomFields []RTIRCustomField `json:"CustomFields"`
}

// DomainRTIRRecord drives upsert into domains + domain_records (ticket_id = RTIR ticket).
type DomainRTIRRecord struct {
	TicketID             string
	Value                string
	Type                 string
	Whitelist            bool
	Description          string
	Tags                 []string
	RecordDate           time.Time
	LastSuccessfulSyncAt time.Time
}
