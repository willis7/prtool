package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var timeNow = time.Now // Allow mocking for tests

// ParseRelativeDuration parses a string like "-7d", "-1m", or "-1yr" into a time.Time.
// It only supports negative durations.
func ParseRelativeDuration(r string) (time.Time, error) {
	if r == "" {
		return time.Time{}, fmt.Errorf("relative duration string cannot be empty")
	}

	if !strings.HasPrefix(r, "-") {
		return time.Time{}, fmt.Errorf("relative duration must be negative (e.g., -7d)")
	}

	now := timeNow()
	// Remove the leading '-'
	r = r[1:]

	var unit string
	var valueStr string

	if strings.HasSuffix(r, "d") {
		unit = "d"
		valueStr = strings.TrimSuffix(r, "d")
	} else if strings.HasSuffix(r, "m") {
		unit = "m"
		valueStr = strings.TrimSuffix(r, "m")
	} else if strings.HasSuffix(r, "yr") {
		unit = "yr"
		valueStr = strings.TrimSuffix(r, "yr")
	} else if strings.HasSuffix(r, "y") {
		unit = "y"
		valueStr = strings.TrimSuffix(r, "y")
	} else {
		return time.Time{}, fmt.Errorf("unknown or missing duration unit in: -%s", r)
	}

	if valueStr == "" {
		return time.Time{}, fmt.Errorf("missing numeric value in duration: -%s", r)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid numeric value in duration: %s", valueStr)
	}

	switch unit {
	case "d":
		return now.AddDate(0, 0, -value), nil
	case "m":
		return now.AddDate(0, -value, 0), nil
	case "y", "yr":
		return now.AddDate(-value, 0, 0), nil
	default:
		// Unreachable
		return time.Time{}, fmt.Errorf("unknown duration unit: %s", unit)
	}
}