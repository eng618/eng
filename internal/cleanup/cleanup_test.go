package cleanup

import (
	"strings"
	"testing"
)

func TestParseReclaimedBytes(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int64
	}{
		{
			name:     "single GB match",
			output:   "Total reclaimed space: 23.15GB\n",
			expected: 23150000000,
		},
		{
			name:     "single MB match",
			output:   "Total: 48.17MB\n",
			expected: 48170000,
		},
		{
			name:     "zero bytes",
			output:   "Total reclaimed space: 0B\n",
			expected: 0,
		},
		{
			name: "multiple matches in output",
			output: "Total reclaimed space: 100MB\n" +
				"Some intermediate logs...\n" +
				"Total: 50MB\n",
			expected: 150000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReclaimedBytes(tt.output)
			if got != tt.expected {
				t.Errorf("ParseReclaimedBytes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReportRenderSummaryTable(t *testing.T) {
	report := &Report{}
	report.Add(ItemResult{
		Name:       "Docker Dangling Layers",
		Category:   "docker",
		Status:     StatusSuccess,
		FreedBytes: 1000 * 1000 * 50, // 50MB
	})
	report.Add(ItemResult{
		Name:       "Docker Unused Images (filter: 168h)",
		Category:   "docker",
		Status:     StatusSuccess,
		FreedBytes: 1000 * 1000 * 1000 * 20, // 20GB
	})
	report.Add(ItemResult{
		Name:     "Homebrew Cache",
		Category: "brew",
		Status:   StatusSkipped,
		Message:  "brew not found",
	})

	output := report.RenderSummaryTable()
	if !strings.Contains(output, "Cleanup Summary") {
		t.Errorf("expected summary title in output, got: %s", output)
	}
	if !strings.Contains(output, "Docker Dangling Layers") {
		t.Errorf("expected operation name in output, got: %s", output)
	}
	if !strings.Contains(output, "Total Disk Space Reclaimed") {
		t.Errorf("expected total row in output, got: %s", output)
	}
}

func TestReportEmpty(t *testing.T) {
	report := &Report{}
	output := report.RenderSummaryTable()
	if output != "" {
		t.Errorf("expected empty output for empty report, got: %s", output)
	}
}
