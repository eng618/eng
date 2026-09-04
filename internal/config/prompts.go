package config

import (
	"errors"
)

// Interactive prompt hooks consumed by this package.
//
// config must never import presentation packages (internal/ui), so all
// user prompting goes through these variables. The command layer wires the
// real implementations at startup (see cmd/config/prompts.go); tests
// override them per-case. Defaults fail loudly so an unwired composition
// surfaces immediately instead of hanging on stdin.
var (
	// ConfirmPrompt asks a yes/no question.
	ConfirmPrompt = confirmUnwired
	// InputPrompt asks for a line of text.
	InputPrompt = inputUnwired
	// SelectPrompt asks to pick one option.
	SelectPrompt = selectUnwired
	// MultiSelectPrompt asks to pick any number of options.
	MultiSelectPrompt = multiSelectUnwired
)

var errPromptsUnwired = errors.New("prompts not wired: command layer must assign prompt hooks")

func confirmUnwired(string, bool) (bool, error) {
	return false, errPromptsUnwired
}

func inputUnwired(string, string) (string, error) {
	return "", errPromptsUnwired
}

func selectUnwired(string, []string, string) (string, error) {
	return "", errPromptsUnwired
}

func multiSelectUnwired(string, []string, []string) ([]string, error) {
	return nil, errPromptsUnwired
}
