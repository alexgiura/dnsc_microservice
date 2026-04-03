package rtir

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseTicketID decodes RTIR "id" when JSON sends either a string or a number.
func ParseTicketID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return strings.TrimSpace(n.String())
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return fmt.Sprintf("%.0f", f)
	}
	return ""
}
