package config

import (
	internalconfig "github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/ui"
)

// Wire the interactive prompt hooks consumed by internal/config.
//
// The config package must never import presentation code, so its prompting
// goes through hook variables with unwired defaults. This init assigns the
// real huh-backed implementations. It runs for every eng invocation since
// cmd/config is always linked via the root command.
func init() {
	internalconfig.ConfirmPrompt = ui.Confirm
	internalconfig.InputPrompt = ui.Input
	internalconfig.SelectPrompt = ui.Select
	internalconfig.MultiSelectPrompt = ui.MultiSelect
}
