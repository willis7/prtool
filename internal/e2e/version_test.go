package e2e

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionFlagReflectsLdflags(t *testing.T) {
	cmd := exec.Command("go", "run", "-ldflags", "-X 'github.com/yourorg/prtool/cmd.version=test-version'", "../../main.go", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("prtool --version failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "test-version") {
		t.Errorf("expected output to contain 'test-version', got: %s", string(out))
	}
}
