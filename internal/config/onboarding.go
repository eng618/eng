package config

import (
	"os"
	"strings"
)

// onboardingSkippedCommands never trigger the first-run setup prompt. They
// either manage configuration themselves (config), only report state
// (doctor, version, logs), render help/completion output that must stay
// clean, or run their own setup flow (system).
var onboardingSkippedCommands = map[string]bool{
	"config":     true,
	"doctor":     true,
	"doc":        true,
	"version":    true,
	"help":       true,
	"completion": true,
	"__complete": true,
	"logs":       true,
	"system":     true,
}

// firstCommand returns the invoked subcommand from raw args, skipping flags
// and the --config value. Returns "" when no subcommand is present.
func firstCommand(args []string) string {
	skipNext := false
	for _, a := range args[1:] {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "--config" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		return strings.ToLower(a)
	}
	return ""
}

// ShouldAutoOnboard decides whether to offer the first-run setup wizard.
// fileCreated reports whether initConfig just created a fresh config file
// (true first run). stdinTTY/stdoutTTY are injected for testability; callers
// probe the real descriptors (e.g. via ui.IsTerminal).
//
// The prompt fires at most once per machine — only on true first run — and
// never for meta/config commands, non-terminals (scripts, CI, pipes), or
// when ENG_NO_ONBOARDING is set. Later runs with missing settings rely on
// actionable command errors pointing at `eng config edit --interactive`.
func ShouldAutoOnboard(args []string, stdinTTY, stdoutTTY, fileCreated bool) bool {
	if os.Getenv("ENG_NO_ONBOARDING") != "" {
		return false
	}
	if !fileCreated {
		return false
	}
	if !stdinTTY || !stdoutTTY {
		return false
	}
	// A pty without terminal capabilities (dumb/unset TERM) cannot drive
	// the huh form reliably; never hang waiting for keys that can't arrive.
	if term := os.Getenv("TERM"); term == "" || term == "dumb" {
		return false
	}
	if cmd := firstCommand(args); cmd == "" || onboardingSkippedCommands[cmd] {
		return false
	}
	return true
}
