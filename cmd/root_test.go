package cmd

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("go", "run", "../main.go", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run: %v, output: %s", err, out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("dev")) {
		t.Errorf("expected output to contain 'dev', got: %s", out.String())
	}
}
