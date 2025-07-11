package timeutil

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var relativeDurationRegex = regexp.MustCompile(`^-(\d+)(d|m|yr)$`)

// ParseRelativeDuration parses a relative duration string like "-7d", "-1m", "-1yr"
// and returns the corresponding time in the past from now.
// Returns an error for empty strings or positive durations.
func ParseRelativeDuration(r string) (time.Time, error) {
	if r == "" {
		return time.Time{}, fmt.Errorf("empty duration string")
	}

	matches := relativeDurationRegex.FindStringSubmatch(r)
	if matches == nil {
		return time.Time{}, fmt.Errorf("invalid duration format: %s", r)
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration amount: %s", matches[1])
	}

	unit := matches[2]
	now := time.Now()

	switch unit {
	case "d":
		return now.AddDate(0, 0, -amount), nil
	case "m":
		return now.AddDate(0, -amount, 0), nil
	case "yr":
		return now.AddDate(-amount, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid duration unit: %s", unit)
	}
}