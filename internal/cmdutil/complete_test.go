package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCompletePrefix(t *testing.T) {
	matches, directive := CompletePrefix([]string{"home", "work", "home-lab"}, "ho")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
	if len(matches) != 2 || matches[0] != "home" || matches[1] != "home-lab" {
		t.Errorf("expected [home home-lab], got %v", matches)
	}

	if matches, _ := CompletePrefix([]string{"a"}, ""); len(matches) != 1 {
		t.Errorf("expected empty prefix to match all, got %v", matches)
	}
	if matches, _ := CompletePrefix([]string{"a"}, "z"); len(matches) != 0 {
		t.Errorf("expected no matches, got %v", matches)
	}
	if matches, _ := CompletePrefix(nil, ""); len(matches) != 0 {
		t.Errorf("expected nil-safe empty result, got %v", matches)
	}
}
