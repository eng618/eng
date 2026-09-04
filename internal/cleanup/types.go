package cleanup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// Status represents the completion state of a cleanup item.
type Status string

const (
	StatusSuccess Status = "Success"
	StatusSkipped Status = "Skipped"
	StatusFailed  Status = "Failed"
)

// ItemResult records the outcome of a single cleanup operation.
type ItemResult struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Status      Status `json:"status"`
	FreedBytes  int64  `json:"freed_bytes"`
	Message     string `json:"message,omitempty"`
	ReclaimText string `json:"reclaim_text,omitempty"`
}

// Report aggregates multiple cleanup results and renders summary metrics.
type Report struct {
	Items           []ItemResult `json:"items"`
	TotalFreedBytes int64        `json:"total_freed_bytes"`
}

// Add appends an item result to the report and updates running totals.
func (r *Report) Add(item ItemResult) {
	r.Items = append(r.Items, item)
	if item.Status == StatusSuccess && item.FreedBytes > 0 {
		r.TotalFreedBytes += item.FreedBytes
	}
}

// RenderSummaryTable formats a terminal summary card for all cleanup results.
func (r *Report) RenderSummaryTable() string {
	if len(r.Items) == 0 {
		return ""
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginBottom(1)

	cellOpStyle := lipgloss.NewStyle().Padding(0, 1)
	cellStatusSuccess := lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true).Padding(0, 1)
	cellStatusSkipped := lipgloss.NewStyle().Foreground(lipgloss.Color("#d97706")).Padding(0, 1)
	cellStatusFailed := lipgloss.NewStyle().Foreground(theme.Destructive).Bold(true).Padding(0, 1)
	cellFreedStyle := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Right)

	var rows []string
	rows = append(rows, lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("  %-45s %-14s %s", "Cleanup Operation", "Status", "Space Reclaimed"),
	))
	rows = append(rows, strings.Repeat("─", 78))

	for _, item := range r.Items {
		var statusText string
		switch item.Status {
		case StatusSuccess:
			statusText = cellStatusSuccess.Render("✓ Success")
		case StatusSkipped:
			statusText = cellStatusSkipped.Render("⚠ Skipped")
		case StatusFailed:
			statusText = cellStatusFailed.Render("✖ Failed")
		default:
			if item.Status == "" {
				statusText = "—"
			} else {
				statusText = string(item.Status)
			}
		}

		freedStr := "—"
		if item.FreedBytes > 0 {
			freedStr = humanize.Bytes(uint64(item.FreedBytes))
		} else if item.ReclaimText != "" {
			freedStr = item.ReclaimText
		}

		opCol := cellOpStyle.Render(fmt.Sprintf("%-43s", ui.Truncate(item.Name, 43)))
		freedCol := cellFreedStyle.Render(freedStr)

		rows = append(rows, fmt.Sprintf("%s %-14s %s", opCol, statusText, freedCol))
	}

	rows = append(rows, strings.Repeat("─", 78))

	totalFormatted := "0 B"
	if r.TotalFreedBytes > 0 {
		totalFormatted = humanize.Bytes(uint64(r.TotalFreedBytes))
	}

	totalStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary)

	rows = append(rows, fmt.Sprintf("  %s %s", totalStyle.Render("Total Disk Space Reclaimed:"), totalStyle.Render(totalFormatted)))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 1).
		MarginTop(1).
		MarginBottom(1).
		Render(strings.Join(rows, "\n"))

	return fmt.Sprintf("\n%s\n%s", headerStyle.Render("🧹 Host & Docker Cleanup Summary"), box)
}

// DockerCleanOptions holds configuration options for Docker cleanup.
type DockerCleanOptions struct {
	All        bool
	OlderThan  string
	BuildCache bool
	Volumes    bool
	DryRun     bool
	Verbose    bool
}

// SystemCleanOptions holds configuration options for full host cleanup.
type SystemCleanOptions struct {
	Docker         bool
	DockerOpts     DockerCleanOptions
	Journal        bool
	JournalSize    string
	Packages       bool
	Asdf           bool
	Brew           bool
	DryRun         bool
	Verbose        bool
	AutoApprove    bool
	CleanupTimeout int
}
