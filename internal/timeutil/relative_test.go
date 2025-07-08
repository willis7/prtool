package timeutil

import (
	"testing"
	"time"
)

func TestParseRelativeDuration(t *testing.T) {
	// Mock time.Now() for consistent testing
	fixedTime := time.Date(2024, time.July, 8, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected time.Time
	}{
		{"7 days ago", "-7d", false, fixedTime.AddDate(0, 0, -7)},
		{"3 days ago", "-3d", false, fixedTime.AddDate(0, 0, -3)},
		{"1 month ago", "-1m", false, fixedTime.AddDate(0, -1, 0)},
		{"3 month ago", "-3m", false, fixedTime.AddDate(0, -3, 0)},
		{"1 year ago", "-1yr", false, fixedTime.AddDate(-1, 0, 0)},
		{"2 years ago", "-2yr", false, fixedTime.AddDate(-2, 0, 0)},
		{"empty string", "", true, time.Time{}},
		{"positive duration", "7d", true, time.Time{}},
		{"invalid number", "-ad", true, time.Time{}},
		{"invalid unit", "-7x", true, time.Time{}},
		{"missing number", "-d", true, time.Time{}},
		{"single year unit", "-1y", false, fixedTime.AddDate(-1, 0, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRelativeDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRelativeDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.expected) {
				t.Errorf("ParseRelativeDuration() got = %v, expected %v", got, tt.expected)
			}
		})
	}
}
