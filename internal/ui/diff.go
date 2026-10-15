package ui

import (
	"strings"

	"github.com/eng618/eng/internal/ui/theme"
)

// RenderDiffLine styles a single unified-diff line for quick review.
// Added lines get a green background, removed lines a red background,
// hunk headers a muted style, and diff metadata a header style.
// The ---/+++ file headers are treated as metadata, not add/del.
// It returns the input unchanged when styling is disabled or the line
// is context.
func RenderDiffLine(line string) string {
	if theme.NoColorEnabled() {
		return line
	}
	switch {
	case strings.HasPrefix(line, "@@"):
		return theme.DiffHunk.Render(line)
	case strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- "):
		return theme.DiffHeader.Render(line)
	case strings.HasPrefix(line, "diff --git ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "new file ") ||
		strings.HasPrefix(line, "old mode ") ||
		strings.HasPrefix(line, "new mode ") ||
		strings.HasPrefix(line, "similarity ") ||
		strings.HasPrefix(line, "rename "):
		return theme.DiffHeader.Render(line)
	case strings.HasPrefix(line, "+"):
		return theme.DiffAddLine.Render(line)
	case strings.HasPrefix(line, "-"):
		return theme.DiffDelLine.Render(line)
	default:
		return line
	}
}

// RenderDiff styles full unified-diff output line by line, preserving
// the trailing newline behavior of the input.
func RenderDiff(output string) string {
	trailing := strings.HasSuffix(output, "\n")
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return output
	}
	for i, line := range lines {
		lines[i] = RenderDiffLine(line)
	}
	rendered := strings.Join(lines, "\n")
	if trailing {
		rendered += "\n"
	}
	return rendered
}

// RenderDiffStat styles diff --stat output by coloring the insertion (+)
// and deletion (-) markers, leaving file names and counts untouched.
func RenderDiffStat(output string) string {
	if theme.NoColorEnabled() {
		return output
	}
	trailing := strings.HasSuffix(output, "\n")
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return output
	}
	for i, line := range lines {
		lines[i] = renderStatLine(line)
	}
	rendered := strings.Join(lines, "\n")
	if trailing {
		rendered += "\n"
	}
	return rendered
}

// renderStatLine colors runs of + and - in a single diffstat line.
func renderStatLine(line string) string {
	var sb strings.Builder
	for _, r := range line {
		switch r {
		case '+':
			sb.WriteString(theme.DiffAddText.Render(string(r)))
		case '-':
			sb.WriteString(theme.DiffDelText.Render(string(r)))
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
