package render

import (
	"fmt"

	"github.com/yourorg/prtool/internal/model"
)

type Metadata struct {
	Title string
	Date  string
}

func Render(meta Metadata, prs []model.PR) string {
	out := fmt.Sprintf("# %s\nGenerated: %s\n\n", meta.Title, meta.Date)
	for _, pr := range prs {
		out += fmt.Sprintf("- [%s](%s) by %s (merged %s)\n", pr.Title, pr.Repo, pr.Author, pr.MergedAt.Format("2006-01-02"))
	}
	return out
}
