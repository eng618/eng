package ui

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{name: "short unchanged", s: "abc", maxLen: 10, expected: "abc"},
		{name: "exact unchanged", s: "abc", maxLen: 3, expected: "abc"},
		{name: "cut with ellipsis", s: "abcdef", maxLen: 5, expected: "ab..."},
		{name: "multibyte safe", s: "héllo wörld", maxLen: 7, expected: "héll..."},
		{name: "tiny limit no ellipsis", s: "abcdef", maxLen: 2, expected: "ab"},
		{name: "zero limit", s: "abcdef", maxLen: 0, expected: ""},
		{name: "negative limit", s: "abcdef", maxLen: -1, expected: ""},
		{name: "empty", s: "", maxLen: 5, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.s, tt.maxLen); got != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.expected)
			}
		})
	}
}

func TestTruncateRuneLength(t *testing.T) {
	got := Truncate(strings.Repeat("é", 10), 5)
	if got != "éé..." {
		t.Errorf("expected rune-safe cut, got %q", got)
	}
	if n := len([]rune(Truncate("abcdef", 5))); n != 5 {
		t.Errorf("expected total length 5 runes, got %d", n)
	}
}
