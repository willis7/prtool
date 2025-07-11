package timeutil

import (
	"strings"
	"testing"
	"time"
)

func TestParseRelativeDuration(t *testing.T) {
	// Fix time for consistent testing
	now := time.Now()

	tests := []struct {
		name      string
		input     string
		wantDays  int
		wantErr   bool
		errContains string
	}{
		// Valid cases
		{
			name:     "7 days ago",
			input:    "-7d",
			wantDays: -7,
			wantErr:  false,
		},
		{
			name:     "1 day ago",
			input:    "-1d",
			wantDays: -1,
			wantErr:  false,
		},
		{
			name:     "30 days ago",
			input:    "-30d",
			wantDays: -30,
			wantErr:  false,
		},
		{
			name:     "1 month ago",
			input:    "-1m",
			wantDays: -30, // Approximate
			wantErr:  false,
		},
		{
			name:     "3 months ago",
			input:    "-3m",
			wantDays: -90, // Approximate
			wantErr:  false,
		},
		{
			name:     "1 year ago",
			input:    "-1yr",
			wantDays: -365, // Approximate
			wantErr:  false,
		},
		{
			name:     "2 years ago",
			input:    "-2yr",
			wantDays: -730, // Approximate
			wantErr:  false,
		},

		// Invalid cases
		{
			name:        "empty string",
			input:       "",
			wantErr:     true,
			errContains: "empty duration string",
		},
		{
			name:        "positive duration not allowed",
			input:       "7d",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "positive with plus sign",
			input:       "+7d",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "invalid unit",
			input:       "-7w",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "invalid format - no number",
			input:       "-d",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "invalid format - no unit",
			input:       "-7",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "invalid format - spaces",
			input:       "- 7d",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "zero duration",
			input:       "-0d",
			wantDays: 0,
			wantErr:  false,
		},
		{
			name:        "decimal not allowed",
			input:       "-1.5d",
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name:        "multiple units not allowed",
			input:       "-1yr2m",
			wantErr:     true,
			errContains: "invalid duration format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRelativeDuration(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRelativeDuration(%q) expected error, got nil", tt.input)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseRelativeDuration(%q) error = %v, want error containing %q", tt.input, err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRelativeDuration(%q) unexpected error = %v", tt.input, err)
				return
			}

			// For valid cases, check the time difference is approximately correct
			// We use a tolerance of ±1 day for month/year calculations due to varying month lengths
			daysDiff := int(now.Sub(got).Hours() / 24)
			tolerance := 1

			// For month and year calculations, allow more tolerance
			if strings.HasSuffix(tt.input, "m") || strings.HasSuffix(tt.input, "yr") {
				tolerance = 5
			}

			expectedDays := -tt.wantDays // Convert to positive for comparison
			if daysDiff < expectedDays-tolerance || daysDiff > expectedDays+tolerance {
				t.Errorf("ParseRelativeDuration(%q) = %v, want approximately %d days ago (got %d days)", 
					tt.input, got, expectedDays, daysDiff)
			}
		})
	}
}

func TestParseRelativeDuration_TimeCalculation(t *testing.T) {
	// Test specific date calculations
	tests := []struct {
		name     string
		input    string
		validate func(result time.Time) bool
	}{
		{
			name:  "days calculation",
			input: "-7d",
			validate: func(result time.Time) bool {
				expected := time.Now().AddDate(0, 0, -7)
				diff := expected.Sub(result).Abs()
				return diff < time.Second
			},
		},
		{
			name:  "months calculation",
			input: "-2m",
			validate: func(result time.Time) bool {
				expected := time.Now().AddDate(0, -2, 0)
				diff := expected.Sub(result).Abs()
				return diff < time.Second
			},
		},
		{
			name:  "years calculation",
			input: "-1yr",
			validate: func(result time.Time) bool {
				expected := time.Now().AddDate(-1, 0, 0)
				diff := expected.Sub(result).Abs()
				return diff < time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRelativeDuration(tt.input)
			if err != nil {
				t.Fatalf("ParseRelativeDuration(%q) unexpected error = %v", tt.input, err)
			}

			if !tt.validate(got) {
				t.Errorf("ParseRelativeDuration(%q) = %v, validation failed", tt.input, got)
			}
		})
	}
}