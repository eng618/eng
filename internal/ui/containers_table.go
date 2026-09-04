package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"

	"github.com/eng618/eng/internal/ui/theme"
)

// StackRow is a presentation DTO for one compose stack row. Callers map
// domain types (e.g. containers.Stack) to rows so this package never
// imports domain packages.
type StackRow struct {
	Name       string
	Containers int
	Status     string
	File       string
}

// PortMapping is a presentation DTO for one published container port.
type PortMapping struct {
	PublishedPort int
	TargetPort    int
}

// ContainerRow is a presentation DTO for one container table row.
type ContainerRow struct {
	Name    string
	Service string
	State   string
	Health  string
	Image   string
	Ports   []PortMapping
}

// GetTerminalWidth returns the current terminal width in columns, defaulting to 100 if unparseable.
func GetTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 100
	}
	return w
}

// RenderStackTable renders a Lip Gloss table showing high-level status for Compose stacks.
func RenderStackTable(stacks []StackRow, termWidth int) string {
	if len(stacks) == 0 {
		return theme.MutedText.Render(
			"No compose stacks found. Check your containers path with 'eng config containers-path'.",
		)
	}
	if termWidth <= 0 {
		termWidth = GetTerminalWidth()
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	// Available width for columns after table borders and inner paddings (4 columns = 5 borders + 8 padding spaces = 13 spaces)
	availWidth := termWidth - 13
	if availWidth < 40 {
		availWidth = 40
	}

	colStack := clamp(availWidth*20/100, 12, 25)
	colCount := 10
	colStatus := 14
	colFile := availWidth - (colStack + colCount + colStatus)
	if colFile < 15 {
		colFile = 15
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("STACK", "CONTAINERS", "STATUS", "COMPOSE FILE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			s := cellStyle
			switch col {
			case 0:
				return s.MaxWidth(colStack)
			case 1:
				return s.Width(colCount).Align(lipgloss.Center)
			case 2:
				return s.MaxWidth(colStatus)
			case 3:
				return s.MaxWidth(colFile)
			}
			return s
		})

	for _, s := range stacks {
		t.Row(
			Truncate(s.Name, colStack),
			strconv.Itoa(s.Containers),
			formatStackStatusBadge(s.Status),
			Truncate(s.File, colFile),
		)
	}

	return t.Render()
}

// RenderContainerTable renders a detailed Lip Gloss table for containers in a specific stack.
func RenderContainerTable(stackName string, containerList []ContainerRow, termWidth int) string {
	if len(containerList) == 0 {
		return theme.MutedText.Render(fmt.Sprintf("No running or registered containers in stack %s.", stackName))
	}

	if termWidth <= 0 {
		termWidth = GetTerminalWidth()
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	// Calculate widths dynamically for 5 columns (6 border chars + 10 padding spaces = 16 spaces overhead)
	overhead := 16
	availWidth := termWidth - overhead
	if availWidth < 50 {
		availWidth = 50
	}

	colStatus := 14
	colService := clamp(availWidth*20/100, 10, 22)
	colName := clamp(availWidth*25/100, 14, 30)

	remWidth := availWidth - (colName + colService + colStatus)
	colPorts := clamp(remWidth*40/100, 12, 28)
	colImage := remWidth - colPorts
	if colImage < 15 {
		colImage = 15
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("CONTAINER", "SERVICE", "STATUS", "PORTS", "IMAGE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			s := cellStyle
			switch col {
			case 0:
				return s.MaxWidth(colName)
			case 1:
				return s.MaxWidth(colService)
			case 2:
				return s.MaxWidth(colStatus)
			case 3:
				return s.MaxWidth(colPorts)
			case 4:
				return s.MaxWidth(colImage)
			}
			return s
		})

	for _, c := range containerList {
		t.Row(
			theme.BoldText.Render(Truncate(c.Name, colName)),
			Truncate(c.Service, colService),
			formatContainerStatusBadge(c.State, c.Health),
			theme.MutedText.Render(formatCompactPublishers(c.Ports, colPorts)),
			theme.MutedText.Render(Truncate(c.Image, colImage)),
		)
	}

	title := theme.PrimaryText.Bold(true).Render(fmt.Sprintf("Stack Details: %s", stackName))
	return fmt.Sprintf("%s\n%s", title, t.Render())
}

func formatStackStatusBadge(status string) string {
	badge := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	lower := strings.ToLower(status)

	switch {
	case strings.HasPrefix(lower, "running"):
		return badge.Background(theme.Secondary).Foreground(theme.Background).Render(" RUNNING ")
	case strings.HasPrefix(lower, "partial"):
		return badge.Background(lipgloss.AdaptiveColor{Light: "#d97706", Dark: "#f59e0b"}).
			Foreground(theme.Background).
			Render(fmt.Sprintf(" %s ", strings.ToUpper(status)))
	case strings.HasPrefix(lower, "stopped"):
		return badge.Background(theme.MutedForeground).Foreground(theme.Background).Render(" STOPPED ")
	default:
		return badge.Background(theme.MutedForeground).
			Foreground(theme.Background).
			Render(fmt.Sprintf(" %s ", strings.ToUpper(status)))
	}
}

func formatContainerStatusBadge(state, health string) string {
	badge := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	st := strings.ToLower(state)
	hl := strings.ToLower(health)

	switch {
	case st == "running" && (hl == "" || hl == "healthy"):
		return badge.Background(theme.Secondary).Foreground(theme.Background).Render(" RUNNING ")
	case hl == "unhealthy":
		return badge.Background(theme.Destructive).Foreground(theme.Background).Render(" UNHEALTHY ")
	case st == "exited":
		return badge.Background(theme.MutedForeground).Foreground(theme.Background).Render(" EXITED ")
	default:
		lbl := strings.ToUpper(state)
		if health != "" {
			lbl += fmt.Sprintf(" (%s)", strings.ToUpper(health))
		}
		return badge.Background(lipgloss.AdaptiveColor{Light: "#d97706", Dark: "#f59e0b"}).
			Foreground(theme.Background).
			Render(fmt.Sprintf(" %s ", lbl))
	}
}

func formatCompactPublishers(pubs []PortMapping, maxColWidth int) string {
	if len(pubs) == 0 {
		return "-"
	}

	var parts []string
	for _, p := range pubs {
		if p.PublishedPort > 0 {
			parts = append(parts, fmt.Sprintf("%d:%d", p.PublishedPort, p.TargetPort))
		} else {
			parts = append(parts, strconv.Itoa(p.TargetPort))
		}
	}

	fullStr := strings.Join(parts, ", ")
	if len(fullStr) <= maxColWidth || len(parts) == 1 {
		return Truncate(fullStr, maxColWidth)
	}

	// Compact multiple ports
	firstPort := parts[0]
	moreCount := len(parts) - 1
	compactStr := fmt.Sprintf("%s (+%d more)", firstPort, moreCount)
	if len(compactStr) <= maxColWidth {
		return compactStr
	}

	return fmt.Sprintf("%d ports", len(parts))
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
