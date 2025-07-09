package render

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/willis7/prtool/internal/model"
)

func TestRender(t *testing.T) {
	// Create a dummy time for consistent output
	dummyTime := time.Date(2024, time.July, 8, 10, 0, 0, 0, time.UTC)

	meta := Metadata{
		Timeframe: "-7d",
		Scope:     "org: test-org",
		PRCount:   2,
		Date:      dummyTime,
	}

	prs := []model.PR{
		{
			Title:    "Feature: Add new awesome feature",
			URL:      "https://github.com/test/repo/pull/1",
			Author:   "dev1",
			MergedAt: dummyTime.Add(-24 * time.Hour),
			Labels:   []string{"feature", "enhancement"},
			Body:     "This PR introduces a new feature that does X and Y.\nIt also fixes a minor bug related to Z.",
		},
		{
			Title:    "Bugfix: Fix critical bug in auth",
			URL:      "https://github.com/test/repo/pull/2",
			Author:   "dev2",
			MergedAt: dummyTime.Add(-48 * time.Hour),
			Labels:   []string{"bug", "security"},
			Body:     "Addresses a critical vulnerability in the authentication module.",
		},
	}

	got := Render(meta, prs, "This is a placeholder for the LLM-generated summary.")

	goldenFile := filepath.Join("testdata", "expected.md")
	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenFile, err)
	}

	if got != string(expected) {
		t.Errorf("Render() mismatch:\n--- GOT ---\n%s\n--- WANT ---\n%s\n", got, string(expected))
	}
}
