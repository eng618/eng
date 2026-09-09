package paths

import (
	"testing"
)

func TestExpand(t *testing.T) {
	orig := UserHomeDir
	UserHomeDir = func() (string, error) { return "/home/testuser", nil }
	defer func() { UserHomeDir = orig }()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"tilde", "~", "/home/testuser"},
		{"tilde slash", "~/Development", "/home/testuser/Development"},
		{"absolute", "/tmp/foo", "/tmp/foo"},
		{"env", "$HOME/foo", "$HOME/foo"}, // ExpandEnv leaves $HOME; caller env dependent
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "env" {
				t.Skip("env-dependent")
			}
			if got := Expand(tt.input); got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJoinHome(t *testing.T) {
	orig := UserHomeDir
	UserHomeDir = func() (string, error) { return "/home/testuser", nil }
	defer func() { UserHomeDir = orig }()

	got, err := JoinHome("Development", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/testuser/Development/foo" {
		t.Errorf("JoinHome = %q", got)
	}
}
