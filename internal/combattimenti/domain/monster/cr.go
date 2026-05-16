package monster

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseCR converts a Challenge Rating display string to its numeric value.
//
// Accepts the fractional ratings used at low CRs ("0", "1/8", "1/4", "1/2")
// and integer ratings beyond that ("1", "2", ... "30"). Detail suffixes such
// as "10 (PE 5.900; BC +4)" are stripped before parsing, so the CR detail
// string stored in the SRD doc round-trips through this function.
//
// Returns an error when the head token cannot be parsed as a CR.
func ParseCR(s string) (float64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0, fmt.Errorf("monster: empty CR")
	}
	// Detail suffix lives after the first whitespace ("10 (PE 5.900...)"),
	// the bare CR is the leading token. Split keeps fractions intact.
	head := strings.Fields(trimmed)[0]

	if strings.Contains(head, "/") {
		parts := strings.SplitN(head, "/", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("monster: malformed fractional CR %q", s)
		}
		num, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, fmt.Errorf("monster: invalid CR numerator in %q: %w", s, err)
		}
		den, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, fmt.Errorf("monster: invalid CR denominator in %q: %w", s, err)
		}
		if den == 0 {
			return 0, fmt.Errorf("monster: zero denominator in %q", s)
		}
		return num / den, nil
	}

	v, err := strconv.ParseFloat(head, 64)
	if err != nil {
		return 0, fmt.Errorf("monster: invalid CR %q: %w", s, err)
	}
	return v, nil
}
