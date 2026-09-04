package system

import (
	"github.com/eng618/eng/internal/cleanup"
	"github.com/eng618/eng/internal/ui"
)

// Wire the task-selection prompt consumed by internal/cleanup.
//
// The cleanup package must never import presentation code for prompting, so
// selection goes through a hook variable with an unwired default. This init
// assigns the real huh-backed implementation. cmd/compose wires the same
// hook for its own clean command; both packages are always linked via root.
func init() {
	cleanup.MultiSelectPrompt = ui.MultiSelect
}
