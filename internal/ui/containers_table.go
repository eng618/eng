package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/ui/theme"
)

// RenderStackTable renders a Lip Gloss table showing high-level status for Compose stacks.
func RenderStackTable(stacks []containers.Stack) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)

	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("STACK", "CONTAINERS", "STATUS", "COMPOSE FILE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		})

	for _, s := range stacks {
		t.Row(
			theme.BoldText.Render(s.Name),
			strconv.Itoa(s.Containers),
			formatStackStatusBadge(s.Status),
			theme.MutedText.Render(s.File),
		)
	}

	return t.Render()
}

// RenderContainerTable renders a detailed Lip Gloss table for containers in a specific stack.
func RenderContainerTable(stackName string, containerList []containers.ContainerDetail) string {
	if len(containerList) == 0 {
		return theme.MutedText.Render(fmt.Sprintf("No running or registered containers in stack %s.", stackName))
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("CONTAINER", "SERVICE", "STATUS", "PORTS", "IMAGE").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		})

	for _, c := range containerList {
		t.Row(
			theme.BoldText.Render(c.Name),
			c.Service,
			formatContainerStatusBadge(c.State, c.Health),
			theme.MutedText.Render(formatPublishers(c.Publishers)),
			theme.MutedText.Render(c.Image),
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
		return badge.Background(lipgloss.Color("#10B981")).Foreground(lipgloss.Color("#000000")).Render(" RUNNING ")
	case strings.HasPrefix(lower, "partial"):
		return badge.Background(lipgloss.Color("#F59E0B")).
			Foreground(lipgloss.Color("#000000")).
			Render(fmt.Sprintf(" %s ", strings.ToUpper(status)))
	case strings.HasPrefix(lower, "stopped"):
		return badge.Background(lipgloss.Color("#6B7280")).Foreground(lipgloss.Color("#FFFFFF")).Render(" STOPPED ")
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
		return badge.Background(lipgloss.Color("#10B981")).Foreground(lipgloss.Color("#000000")).Render(" RUNNING ")
	case hl == "unhealthy":
		return badge.Background(lipgloss.Color("#EF4444")).Foreground(lipgloss.Color("#FFFFFF")).Render(" UNHEALTHY ")
	case st == "exited":
		return badge.Background(lipgloss.Color("#6B7280")).Foreground(lipgloss.Color("#FFFFFF")).Render(" EXITED ")
	default:
		lbl := strings.ToUpper(state)
		if health != "" {
			lbl += fmt.Sprintf(" (%s)", strings.ToUpper(health))
		}
		return badge.Background(lipgloss.Color("#F59E0B")).
			Foreground(lipgloss.Color("#000000")).
			Render(fmt.Sprintf(" %s ", lbl))
	}
}

func formatPublishers(pubs []containers.Publisher) string {
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
	return strings.Join(parts, ", ")
}
