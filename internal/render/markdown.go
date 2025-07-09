package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/willis7/prtool/internal/model"
)

// Metadata holds information about the PR summary.
type Metadata struct {
	Timeframe string
	Scope     string
	PRCount   int
	Date      time.Time
}

// Render generates a Markdown string from metadata, PRs, and LLM summary.
func Render(meta Metadata, prs []model.PR, summary string) string {
	var builder strings.Builder

	// Metadata block
	builder.WriteString(fmt.Sprintf("---\nTimeframe: %s\nScope: %s\nPRCount: %d\nDate: %s\n---\n\n", meta.Timeframe, meta.Scope, meta.PRCount, meta.Date.Format("2006-01-02")))

	// LLM summary
	builder.WriteString("## Pull Request Summary\n\n")
	builder.WriteString(summary + "\n\n")

	// List of PRs (for dry-run or debugging)
	if len(prs) > 0 {
		builder.WriteString("### Individual Pull Requests\n\n")
		for _, pr := range prs {
			builder.WriteString(fmt.Sprintf("- [%s](%s) by %s (Merged: %s)\n", pr.Title, pr.URL, pr.Author, pr.MergedAt.Format("2006-01-02")))
			if pr.Body != "" {
				// Indent body for better readability in Markdown list
				indentedBody := strings.ReplaceAll(pr.Body, "\n", "\n  ")
				builder.WriteString(fmt.Sprintf("  %s\n", indentedBody))
			}
			if len(pr.Labels) > 0 {
				builder.WriteString(fmt.Sprintf("  Labels: %s\n", strings.Join(pr.Labels, ", ")))
			}
		}
	}

	return builder.String()
}
