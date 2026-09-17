package auth

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// showCmd prints effective GitLab auth/config details without exposing secrets.
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show GitLab auth and defaults (no secrets)",
	RunE: func(cmd *cobra.Command, args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🔐 GitLab Authentication Details"))
		}

		host := viper.GetString("gitlab.host")
		project := viper.GetString("gitlab.project")

		// Determine token source availability without exposing secrets.
		tokenSource := "none"
		if _, source, err := config.ResolveGitLabToken(); err == nil {
			tokenSource = source
		} else if ref := config.GitLabTokenRef(viper.GetViper()); ref.Provider != config.ProviderEnv {
			tokenSource = fmt.Sprintf("%s:%s (not found)", ref.Provider, ref.Item)
		}

		log.Message("GitLab defaults:")
		log.Message("  host:    %s", valueOrDash(host))
		log.Message("  project: %s", valueOrDash(project))
		log.Message("  token:   %s", tokenSource)
		return nil
	},
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func init() {
	AuthCmd.AddCommand(showCmd)
}
