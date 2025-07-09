package timeutil

import (
	"testing"
	"time"
)

func TestParseRelativeDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkResult func(result time.Time, input string) error
	}{
		// Valid cases
		{
			name:        "7 days ago",
			input:       "-7d",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				// Should be 7 days before now
				expected := time.Now().AddDate(0, 0, -7)
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				// Allow some tolerance for test execution time
				diff := expected.Sub(result)
				if diff > time.Minute || diff < -time.Minute {
					return &testError{"result not approximately 7 days ago"}
				}
				return nil
			},
		},
		{
			name:        "1 month ago",
			input:       "-1m",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
		{
			name:        "1 year ago",
			input:       "-1yr",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
		{
			name:        "30 days ago",
			input:       "-30d",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
		{
			name:        "2 weeks ago",
			input:       "-2w",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
		{
			name:        "12 hours ago",
			input:       "-12h",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
		{
			name:        "alternative year format",
			input:       "-2year",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
		{
			name:        "alternative month format",
			input:       "-6months",
			expectError: false,
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},

		// Error cases
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "positive duration not allowed",
			input:       "7d",
			expectError: true,
		},
		{
			name:        "positive duration with plus",
			input:       "+7d",
			expectError: true,
		},
		{
			name:        "invalid format - no number",
			input:       "-d",
			expectError: true,
		},
		{
			name:        "invalid format - no unit",
			input:       "-7",
			expectError: true,
		},
		{
			name:        "invalid format - letters in number",
			input:       "-a7d",
			expectError: true,
		},
		{
			name:        "zero duration",
			input:       "-0d",
			expectError: true,
		},
		{
			name:        "unsupported unit",
			input:       "-7x",
			expectError: true,
		},
		{
			name:        "invalid format - spaces",
			input:       "- 7d",
			expectError: true,
		},
		{
			name:        "invalid format - mixed case issues",
			input:       "-7D",
			expectError: false, // Should work - case insensitive
			checkResult: func(result time.Time, input string) error {
				if result.After(time.Now()) {
					return &testError{"result should be in the past"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRelativeDuration(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				return
			}

			if tt.checkResult != nil {
				if checkErr := tt.checkResult(result, tt.input); checkErr != nil {
					t.Errorf("Result check failed for input %q: %v", tt.input, checkErr)
				}
			}
		})
	}
}

func TestParseRelativeDurationSpecificCases(t *testing.T) {
	// Test specific parsing behavior
	tests := []struct {
		name     string
		input    string
		expected func() time.Time
	}{
		{
			name:  "exactly 1 day ago",
			input: "-1d",
			expected: func() time.Time {
				return time.Now().AddDate(0, 0, -1)
			},
		},
		{
			name:  "exactly 1 hour ago",
			input: "-1h",
			expected: func() time.Time {
				return time.Now().Add(-1 * time.Hour)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			result, err := ParseRelativeDuration(tt.input)
			after := time.Now()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			expected := tt.expected()

			// Result should be between what we expected calculated before and after the call
			expectedBefore := expected.Add(-time.Duration(after.Sub(before)))
			expectedAfter := expected

			if result.Before(expectedBefore) || result.After(expectedAfter) {
				t.Errorf("Result %v not in expected range [%v, %v]", result, expectedBefore, expectedAfter)
			}
		})
	}
}

func TestParseRelativeDurationUnits(t *testing.T) {
	// Test all supported units
	unitTests := []struct {
		unit        string
		input       string
		expectError bool
	}{
		// Days
		{"d", "-1d", false},
		{"day", "-1day", false},
		{"days", "-1days", false},

		// Weeks
		{"w", "-1w", false},
		{"week", "-1week", false},
		{"weeks", "-1weeks", false},

		// Months
		{"m", "-1m", false},
		{"month", "-1month", false},
		{"months", "-1months", false},

		// Years
		{"y", "-1y", false},
		{"yr", "-1yr", false},
		{"year", "-1year", false},
		{"years", "-1years", false},

		// Hours
		{"h", "-1h", false},
		{"hour", "-1hour", false},
		{"hours", "-1hours", false},

		// Minutes
		{"min", "-1min", false},
		{"minute", "-1minute", false},
		{"minutes", "-1minutes", false},

		// Seconds
		{"s", "-1s", false},
		{"sec", "-1sec", false},
		{"second", "-1second", false},
		{"seconds", "-1seconds", false},

		// Invalid units
		{"invalid", "-1invalid", true},
		{"x", "-1x", true},
	}

	for _, tt := range unitTests {
		t.Run(tt.unit, func(t *testing.T) {
			_, err := ParseRelativeDuration(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for unit %q, but got none", tt.unit)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for unit %q: %v", tt.unit, err)
			}
		})
	}
}

// Custom error type for test result checking
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
