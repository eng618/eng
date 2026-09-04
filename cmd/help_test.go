package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestHelpContainsNoLocalPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || home == "/" {
		t.Skip("no usable home dir")
	}
	var check func(c *cobra.Command)
	check = func(c *cobra.Command) {
		var sb strings.Builder
		c.SetOut(&sb)
		c.SetErr(&sb)
		_ = c.Help()
		if strings.Contains(sb.String(), home) {
			t.Errorf("help for %q embeds local home dir %q", c.CommandPath(), home)
		}
		for _, sub := range c.Commands() {
			check(sub)
		}
	}
	check(GetRootCommand())
}
