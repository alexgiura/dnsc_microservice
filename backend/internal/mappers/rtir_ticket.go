package mappers

import (
	"fmt"
	"strings"
	"time"

	"dnsc_microservice/internal/models"
)

const (
	cfIOCDomains          = "IOC Domains"
	cfPrimaryIncidentType = "Primary Incident Type"
	cfRelatedIncidentType = "Related Incident Type"
)

// RTIRExtractedData is ticket CustomFields mapped to our domain/record fields (one row per IOC domain value).
type RTIRExtractedData struct {
	Domains     []string
	Description string
	Tags        []string
	RecordTime  time.Time
}

// RecordTimeFromRTIRTicket matches core.domain_records.date: RFC3339 Created, else LastUpdated, else UTC now.
func RecordTimeFromRTIRTicket(t *models.RTIRTicketDetail) time.Time {
	if t == nil {
		return time.Now().UTC()
	}
	if ts, ok := parseRTIRTicketTime(t.Created); ok {
		return ts
	}
	if ts, ok := parseRTIRTicketTime(t.LastUpdated); ok {
		return ts
	}
	return time.Now().UTC()
}

func parseRTIRTicketTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

// ExtractRTIRTicketData maps RTIR CustomFields into internal fields. Domains lists each IOC separately (deduped).
func ExtractRTIRTicketData(t *models.RTIRTicketDetail) (RTIRExtractedData, error) {
	var out RTIRExtractedData
	if t == nil {
		return out, fmt.Errorf("nil ticket")
	}
	out.RecordTime = RecordTimeFromRTIRTicket(t)

	var ioc, primary, related []string
	for _, cf := range t.CustomFields {
		switch strings.TrimSpace(cf.Name) {
		case cfIOCDomains:
			ioc = append(ioc, cf.Values...)
		case cfPrimaryIncidentType:
			primary = append(primary, cf.Values...)
		case cfRelatedIncidentType:
			related = append(related, cf.Values...)
		}
	}

	seen := make(map[string]struct{})
	for _, s := range ioc {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out.Domains = append(out.Domains, s)
	}

	if len(primary) > 0 {
		parts := make([]string, 0, len(primary))
		for _, s := range primary {
			s = strings.TrimSpace(s)
			if s != "" {
				parts = append(parts, s)
			}
		}
		out.Description = strings.Join(parts, ", ")
	}
	if strings.TrimSpace(out.Description) == "" {
		out.Description = "Phishing"
	}

	for _, s := range related {
		s = strings.TrimSpace(s)
		if s != "" {
			out.Tags = append(out.Tags, s)
		}
	}
	return out, nil
}
