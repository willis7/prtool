package timeutil

import (
	"testing"
	"time"
)

func TestParseRelativeDuration_Valid(t *testing.T) {
	now := time.Now()
	tests := []struct {
		input string
		check func(time.Time) bool
	}{
		{"-7d", func(ts time.Time) bool { return ts.Before(now) }},
		{"-1m", func(ts time.Time) bool { return ts.Before(now) }},
		{"-1yr", func(ts time.Time) bool { return ts.Before(now) }},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out, err := ParseRelativeDuration(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.check(out) {
				t.Errorf("expected output before now, got %v", out)
			}
		})
	}
}

func TestParseRelativeDuration_Invalid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"7d"},
		{"-1w"},
		{"-0d"},
		{"-1"},
		{"-1x"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseRelativeDuration(tt.input)
			if err == nil {
				t.Errorf("expected error for input %q", tt.input)
			}
		})
	}
}
