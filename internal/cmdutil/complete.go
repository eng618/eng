package cmdutil

import (
	"strings"

	"github.com/spf13/cobra"
)

// CompletePrefix filters candidates by the given prefix for shell
// completion and disables file completion. This is the single shared helper
// for ValidArgsFunction and RegisterFlagCompletionFunc implementations;
// do not reimplement prefix filtering per command.
func CompletePrefix(candidates []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var matches []string
	for _, c := range candidates {
		if strings.HasPrefix(c, toComplete) {
			matches = append(matches, c)
		}
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}
