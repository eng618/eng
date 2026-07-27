package asdf

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// MissingRequirement represents a version required by a .tool-versions file that is not installed.
type MissingRequirement struct {
	Plugin     string
	Version    string
	SourceFile string
}

// CheckMissingRequirements compares requirements across all discovered .tool-versions files
// against installed versions, returning any missing requirements.
func CheckMissingRequirements(summaries []FileProtectionSummary, installed map[string][]string) []MissingRequirement {
	var missing []MissingRequirement
	seen := make(map[string]bool)

	for _, summary := range summaries {
		for plugin, reqVersions := range summary.Protected {
			installedVersions := installed[plugin]

			for _, ver := range reqVersions {
				if !slices.Contains(installedVersions, ver) {
					key := plugin + "@" + ver + "@" + summary.FilePath
					if !seen[key] {
						seen[key] = true
						missing = append(missing, MissingRequirement{
							Plugin:     plugin,
							Version:    ver,
							SourceFile: summary.FilePath,
						})
					}
				}
			}
		}
	}

	return missing
}

// FormatSource returns a user-friendly display path for the source file.
func (m MissingRequirement) FormatSource(homeDir string) string {
	if homeDir != "" && strings.HasPrefix(m.SourceFile, homeDir) {
		return "~" + strings.TrimPrefix(m.SourceFile, homeDir)
	}
	return m.SourceFile
}

// WriteToolVersions updates or rewrites a .tool-versions file with the specified ToolVersions map.
func WriteToolVersions(filePath string, tv ToolVersions) error {
	var lines []string

	// Read existing file to preserve comments if possible
	if file, err := os.Open(filePath); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				lines = append(lines, line)
			}
		}
		_ = file.Close()
	}

	// Write new plugin lines
	for plugin, versions := range tv {
		if len(versions) > 0 {
			lines = append(lines, fmt.Sprintf("%s %s", plugin, strings.Join(versions, " ")))
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filePath, []byte(content), 0644)
}
