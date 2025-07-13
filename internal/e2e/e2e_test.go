package e2e

import (
	"os/exec"
	"strings"
	"testing"
)

func TestPrtoolRun_StubLLM_MockGitHub(t *testing.T) {
	cmd := exec.Command("go", "run", "../../main.go", "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("prtool run failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "This is a stub summary.") {
		t.Errorf("expected stub summary in output, got: %s", string(out))
	}
}
