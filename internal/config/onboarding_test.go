package config

import (
	"os"
	"testing"
)

func TestShouldAutoOnboard(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		stdinTTY        bool
		stdoutTTY       bool
		fileCreated     bool
		envNoOnboarding bool
		term            string
		unsetTerm       bool
		want            bool
	}{
		{
			name:        "first run interactive git",
			args:        []string{"eng", "git", "sync-all"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        true,
		},
		{
			name:        "not first run",
			args:        []string{"eng", "git", "sync-all"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: false,
			want:        false,
		},
		{
			name:        "piped stdin",
			args:        []string{"eng", "git"},
			stdinTTY:    false,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "piped stdout",
			args:        []string{"eng", "git"},
			stdinTTY:    true,
			stdoutTTY:   false,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "no subcommand",
			args:        []string{"eng"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "config skipped",
			args:        []string{"eng", "config", "edit"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "doctor skipped",
			args:        []string{"eng", "doctor"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "doc alias skipped",
			args:        []string{"eng", "doc"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "system skipped",
			args:        []string{"eng", "system", "setup"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "version skipped",
			args:        []string{"eng", "version"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "help skipped",
			args:        []string{"eng", "--help"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "complete skipped",
			args:        []string{"eng", "__complete", "git", ""},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:        "dashboard prompts",
			args:        []string{"eng", "dashboard"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        true,
		},
		{
			name:        "global flags before command",
			args:        []string{"eng", "-v", "git"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        true,
		},
		{
			name:        "config flag value skipped",
			args:        []string{"eng", "--config", "/tmp/x.yaml", "doctor"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			want:        false,
		},
		{
			name:            "env opt-out",
			args:            []string{"eng", "git"},
			stdinTTY:        true,
			stdoutTTY:       true,
			fileCreated:     true,
			envNoOnboarding: true,
			want:            false,
		},
		{
			name:        "dumb terminal",
			args:        []string{"eng", "git"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			term:        "dumb",
			want:        false,
		},
		{
			name:        "unset terminal",
			args:        []string{"eng", "git"},
			stdinTTY:    true,
			stdoutTTY:   true,
			fileCreated: true,
			unsetTerm:   true,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envNoOnboarding {
				t.Setenv("ENG_NO_ONBOARDING", "1")
			}
			if tt.unsetTerm {
				old, hadOld := os.LookupEnv("TERM")
				if err := os.Unsetenv("TERM"); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if hadOld {
						_ = os.Setenv("TERM", old)
					}
				})
			} else {
				term := tt.term
				if term == "" {
					term = "xterm-256color"
				}
				t.Setenv("TERM", term)
			}
			if got := ShouldAutoOnboard(tt.args, tt.stdinTTY, tt.stdoutTTY, tt.fileCreated); got != tt.want {
				t.Errorf("ShouldAutoOnboard(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
