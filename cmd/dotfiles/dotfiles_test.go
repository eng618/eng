package dotfiles

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils/log"
)

func TestDotfilesCmd_InfoMissingKeys(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()

	// Ensure help output is written to our buffer
	DotfilesCmd.SetOut(&buf)
	DotfilesCmd.SetErr(&buf)
	// Run the command as if DotfilesCmd was invoked with no args
	DotfilesCmd.Run(DotfilesCmd, []string{})

	out := buf.String()
	if out == "" {
		t.Fatalf("expected help output but got empty output")
	}
}

func TestDotfilesCmd_InfoShowsConfig(t *testing.T) {
	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")

	// Ensure DotfilesCmd writes help/info to our buffer
	DotfilesCmd.SetOut(&buf)
	DotfilesCmd.SetErr(&buf)
	// Set the info flag on the command and record prior value to restore
	if err := DotfilesCmd.Flags().Set("info", "true"); err != nil {
		t.Fatalf("failed to set info flag: %v", err)
	}
	defer func() {
		_ = DotfilesCmd.Flags().Set("info", "false")
	}()

	DotfilesCmd.Run(DotfilesCmd, []string{})

	out := buf.String()
	if !strings.Contains(out, "Repository Path:") {
		t.Fatalf("repo path not shown: %s", out)
	}
	if !strings.Contains(out, "/tmp/repo") {
		t.Fatalf("repo path value not shown: %s", out)
	}
}
