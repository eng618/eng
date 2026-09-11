package ui

import (
	"strings"
	"testing"
)

func TestRenderKeyValueTable(t *testing.T) {
	rows := []KeyValueRow{
		{Label: "Hostname", Value: "testhost"},
		{Label: "CPU", Value: "Fake CPU 9000 (4 cores)"},
	}

	out := RenderKeyValueTable("System Information", rows, 100)
	if !strings.Contains(out, "FIELD") || !strings.Contains(out, "VALUE") {
		t.Errorf("Expected table headers, got:\n%s", out)
	}
	if !strings.Contains(out, "Hostname") || !strings.Contains(out, "testhost") {
		t.Errorf("Expected row content, got:\n%s", out)
	}
	if !strings.Contains(out, "System Information") {
		t.Errorf("Expected title, got:\n%s", out)
	}

	// Empty title renders the bare table.
	bare := RenderKeyValueTable("", rows, 100)
	if strings.Contains(bare, "System Information") {
		t.Errorf("Expected no title, got:\n%s", bare)
	}

	// Narrow terminals still render intact borders.
	for _, w := range []int{0, 40, 60, 120} {
		narrow := RenderKeyValueTable("", rows, w)
		for line := range strings.Lines(narrow) {
			if strings.HasPrefix(line, "┌") && !strings.HasSuffix(strings.TrimSpace(line), "┐") {
				t.Errorf("scrambled top border for width %d: %s", w, line)
			}
			if strings.HasPrefix(line, "└") && !strings.HasSuffix(strings.TrimSpace(line), "┘") {
				t.Errorf("scrambled bottom border for width %d: %s", w, line)
			}
		}
	}
}

func TestRenderKeyValueTableEmpty(t *testing.T) {
	out := RenderKeyValueTable("System Information", nil, 100)
	if !strings.Contains(out, "No system information available.") {
		t.Errorf("Expected empty message, got:\n%s", out)
	}
}
