package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestInitCmd_GeneratesConfig(t *testing.T) {
	_ = os.Remove("prtool.config.yaml") // cleanup before
	root := &cobra.Command{Use: "prtool"}
	root.AddCommand(initCmd)
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatalf("initCmd failed: %v", err)
	}
	info, err := os.Stat("prtool.config.yaml")
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("config file is empty")
	}
	_ = os.Remove("prtool.config.yaml") // cleanup after
}
