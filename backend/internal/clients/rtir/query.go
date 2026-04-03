package rtir

import (
	"fmt"
	"time"
)

// BuildSearchQuery builds the RTIR tickets search query string (Queue + blacklist + Updated).
func BuildSearchQuery(updatedAfter time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}
	s := updatedAfter.In(loc).Format("2006-01-02 15:04:05")
	return fmt.Sprintf(
		`(Queue='SOC' OR Queue='PNRISC-SOC') AND 'CF.{blacklist}'='yes' AND Updated > '%s'`,
		s,
	)
}
