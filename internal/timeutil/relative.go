package timeutil

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var re = regexp.MustCompile(`^-(\d+)([a-z]+)$`)

func ParseRelativeDuration(r string) (time.Time, error) {
	if r == "" {
		return time.Time{}, errors.New("empty string")
	}
	if !strings.HasPrefix(r, "-") {
		return time.Time{}, errors.New("positive durations not allowed")
	}
	m := re.FindStringSubmatch(r)
	if len(m) != 3 {
		return time.Time{}, errors.New("invalid format")
	}
	val, err := strconv.Atoi(m[1])
	if err != nil {
		return time.Time{}, err
	}
	if val == 0 {
		return time.Time{}, errors.New("zero value not allowed")
	}

	now := time.Now()
	switch m[2] {
	case "d":
		return now.AddDate(0, 0, -val), nil
	case "m":
		return now.AddDate(0, -val, 0), nil
	case "yr":
		return now.AddDate(-val, 0, 0), nil
	default:
		return time.Time{}, errors.New("unsupported unit")
	}
}
