package render

import (
	"testing"

	"github.com/yourorg/prtool/internal/model"
)

func TestRender_Markdown(t *testing.T) {
	meta := Metadata{Title: "PRs", Date: "2025-07-13"}
	prs := []model.PR{{Title: "Fix bug", Author: "alice", Repo: "repo", MergedAt: model.PR{}.MergedAt}}
	md := Render(meta, prs)
	if md == "" || md[0] != '#' {
		t.Errorf("unexpected markdown output: %q", md)
	}
	if want := "# PRs"; md[:len(want)] != want {
		t.Errorf("expected header, got %q", md)
	}
}
