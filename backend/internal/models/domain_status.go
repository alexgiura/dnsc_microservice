package models

import (
	"fmt"
	"strings"
)

// Domain list classification (core.domains.status).
const (
	DomainStatusWhitelist = "whitelist"
	DomainStatusBlacklist = "blacklist"
	DomainStatusPending   = "pending"
	DomainStatusRejected  = "rejected"
)

// ParseDomainStatus normalizes and validates an API / DB status string.
func ParseDomainStatus(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case DomainStatusWhitelist, DomainStatusBlacklist, DomainStatusPending, DomainStatusRejected:
		return s, nil
	default:
		return "", fmt.Errorf("invalid domain status %q (expected whitelist, blacklist, pending, rejected)", s)
	}
}
