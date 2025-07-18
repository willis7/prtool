package timeutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseRelativeDuration parses relative duration strings like "-7d", "-1m", "-1yr"
// and returns the corresponding time.Time relative to now.
// Only negative durations (past times) are allowed.
func ParseRelativeDuration(r string) (time.Time, error) {
	if r == "" {
		return time.Time{}, fmt.Errorf("duration string cannot be empty")
	}

	// Must start with minus sign (only past times allowed)
	if !strings.HasPrefix(r, "-") {
		return time.Time{}, fmt.Errorf("duration must be negative (past time): %s", r)
	}

	// Remove the minus sign for parsing
	durationStr := r[1:]

	// Regular expression to match number followed by unit
	re := regexp.MustCompile(`^(\d+)([a-zA-Z]+)$`)
	matches := re.FindStringSubmatch(durationStr)

	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("invalid duration format: %s (expected format: -<number><unit>)", r)
	}

	// Parse the number
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number in duration: %s", matches[1])
	}

	if num <= 0 {
		return time.Time{}, fmt.Errorf("duration number must be positive: %d", num)
	}

	// Parse the unit
	unit := strings.ToLower(matches[2])
	now := time.Now()

	switch unit {
	case "d", "day", "days":
		return now.AddDate(0, 0, -num), nil
	case "w", "week", "weeks":
		return now.AddDate(0, 0, -num*7), nil
	case "m", "month", "months":
		return now.AddDate(0, -num, 0), nil
	case "y", "yr", "year", "years":
		return now.AddDate(-num, 0, 0), nil
	case "h", "hour", "hours":
		return now.Add(-time.Duration(num) * time.Hour), nil
	case "min", "minute", "minutes":
		return now.Add(-time.Duration(num) * time.Minute), nil
	case "s", "sec", "second", "seconds":
		return now.Add(-time.Duration(num) * time.Second), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time unit: %s (supported: d, w, m, y, h, min, s)", unit)
	}
}
