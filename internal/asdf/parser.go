package asdf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseToolVersionsFile reads a .tool-versions file from the given path
// and returns a ToolVersions map of plugin -> protected versions.
func ParseToolVersionsFile(filePath string) (ToolVersions, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tool-versions file %s: %w", filePath, err)
	}
	defer file.Close()

	return ParseToolVersions(file)
}

// ParseToolVersions reads from an io.Reader (such as a .tool-versions file)
// and returns a ToolVersions map of plugin -> protected versions.
func ParseToolVersions(r io.Reader) (ToolVersions, error) {
	result := make(ToolVersions)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		plugin := fields[0]
		versions := fields[1:]

		result[plugin] = append(result[plugin], versions...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading tool-versions stream: %w", err)
	}

	return result, nil
}

// ParseASDFListOutput parses the text output of `asdf list` into a map
// of plugin names to slices of installed versions.
func ParseASDFListOutput(output string) (map[string][]string, error) {
	installed := make(map[string][]string)
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentPlugin string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if line starts with whitespace (indicating a version line under current plugin)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if currentPlugin == "" {
				continue
			}
			// Clean version string: remove leading '*' and whitespace
			v := strings.TrimPrefix(trimmed, "*")
			v = strings.TrimSpace(v)
			if v != "" {
				installed[currentPlugin] = append(installed[currentPlugin], v)
			}
		} else {
			// Top-level plugin header line
			currentPlugin = trimmed
			if _, exists := installed[currentPlugin]; !exists {
				installed[currentPlugin] = []string{}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning asdf list output: %w", err)
	}

	return installed, nil
}
