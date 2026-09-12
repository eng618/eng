// Package editor centralizes default-editor detection and resolution.
//
// It mirrors the `ide()` shell helper precedence (agy-ide > code > nano)
// while honoring an explicit `git.editor` config value and the
// $VISUAL/$EDITOR environment variables. Both the dashboard open logic
// and the `eng config git-editor` setter consume this package so the
// available-editor list stays in one place.
package editor
