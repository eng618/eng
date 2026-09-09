package immich

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dustin/go-humanize"

	"github.com/eng618/eng/internal/ui/theme"
)

// RenderStatus renders a comprehensive formatted status card.
func RenderStatus(s *StatusResult, termWidth int) string {
	_ = termWidth // reserved for future table-width support; currently unused

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(theme.Primary).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	borderStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	var sb strings.Builder

	// Title
	title := theme.PrimaryText.Bold(true).Render("📸 Immich Photo Stack Health & Metrics")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Host check note if not configured locally
	if !s.IsHostHostable {
		notice := theme.WarningBanner.Render("NOTE") + " " +
			theme.MutedText.Render(
				fmt.Sprintf(
					"Immich stack is not configured on this host (%s). Probing remote/network API only.",
					s.HostOS,
				),
			)
		sb.WriteString(notice)
		sb.WriteString("\n\n")
	}

	// Overview Table
	tOverview := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers("COMPONENT", "STATE / VALUE", "DETAILS").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		})

	// Service state badge
	svcBadge := formatBadge(s.Service.ActiveState, s.Service.ActiveState == "active")
	timerBadge := theme.MutedText.Render(s.Timer.Left)
	if s.Timer.NextTrigger != "" {
		timerBadge = fmt.Sprintf("Next: %s (%s)", s.Timer.NextTrigger, s.Timer.Left)
	}

	apiBadge := formatBadge("ONLINE", s.API.Reachable)
	if !s.API.Reachable {
		apiBadge = formatBadge("OFFLINE", false)
	}
	apiDetails := fmt.Sprintf("%.2fms latency (%s)", s.API.LatencyMs, s.API.URL)

	dbDetails := fmt.Sprintf("Users: %d | Assets: %s | Albums: %d (Storage: %s)",
		s.Database.Users, humanize.Comma(int64(s.Database.Assets)), s.Database.Albums, s.Database.StorageDir)
	if !s.IsHostHostable {
		dbDetails = "Local database storage not present on this host"
	}

	backupDetails := fmt.Sprintf("Latest: %s (%s) | Total Snapshots: %d",
		s.Backup.LatestDB, s.Backup.LatestSize, s.Backup.TotalBackups)
	if s.Backup.LatestDB == "" {
		backupDetails = "No database backups found on local storage"
	}

	tOverview.Row("Systemd Service", svcBadge, s.Service.Name+" ("+s.Service.UnitFileState+")")
	tOverview.Row("Backup Timer", s.Timer.Name, timerBadge)
	tOverview.Row("Immich API", apiBadge, apiDetails)
	tOverview.Row("Postgres Database", fmt.Sprintf("%d Tables", s.Database.Tables), dbDetails)
	tOverview.Row("Backup Storage", s.Backup.Destination, backupDetails)

	sb.WriteString(tOverview.Render())
	sb.WriteString("\n\n")

	// Containers Table
	if len(s.Containers) > 0 {
		cTable := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(borderStyle).
			Headers("CONTAINER", "SERVICE", "STATUS", "HEALTH").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return headerStyle
				}
				return cellStyle
			})

		for _, c := range s.Containers {
			cTable.Row(
				theme.BoldText.Render(c.Name),
				c.Service,
				formatBadge(c.State, c.State == "running"),
				formatBadge(c.Health, c.Health == "healthy" || c.Health == ""),
			)
		}
		sb.WriteString(theme.PrimaryText.Bold(true).Render("Containers:"))
		sb.WriteString("\n")
		sb.WriteString(cTable.Render())
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatBadge(text string, isOk bool) string {
	badge := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if isOk {
		return badge.Background(lipgloss.Color("#10B981")).
			Foreground(lipgloss.Color("#000000")).
			Render(" " + strings.ToUpper(text) + " ")
	}
	return badge.Background(lipgloss.Color("#EF4444")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render(" " + strings.ToUpper(text) + " ")
}
