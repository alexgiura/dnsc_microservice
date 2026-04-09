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
)

// ParseDomainStatus normalizes and validates an API / DB status string.
func ParseDomainStatus(s string) (string, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case DomainStatusWhitelist, DomainStatusBlacklist, DomainStatusPending:
		return s, nil
	default:
		return "", fmt.Errorf("invalid domain status %q (expected whitelist, blacklist, pending)", s)
	}
}
