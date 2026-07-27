package asdf

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

// ToolVersions maps plugin names to their protected version strings as defined in .tool-versions.
type ToolVersions map[string][]string

// CleanupTarget represents an installed plugin version eligible for removal.
type CleanupTarget struct {
	Plugin    string
	Version   string
	SizeBytes int64
}

// FormatSize returns a human-readable string representation of SizeBytes (e.g., "12.4 MB").
func (c CleanupTarget) FormatSize() string {
	if c.SizeBytes <= 0 {
		return "0 B"
	}
	return humanize.Bytes(uint64(c.SizeBytes))
}

// FormatTargetLabel returns a user-friendly label for a CleanupTarget including disk size,
// e.g. "nodejs @ 20.19.5 (112.4 MB)".
func (c CleanupTarget) FormatTargetLabel() string {
	if c.SizeBytes > 0 {
		return fmt.Sprintf("%s @ %s (%s)", c.Plugin, c.Version, c.FormatSize())
	}
	return fmt.Sprintf("%s @ %s", c.Plugin, c.Version)
}
