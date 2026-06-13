package cmd

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/ui/dashboard"
)

// dashboardCmd represents the dashboard command.
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch the interactive Project & Git Dashboard",
	Long: `Opens a full-screen "mission control" interface to view all configured projects.
It displays your projects in a list, and shows the live status of their repositories
(cloned state, current branch, and uncommitted changes) in real-time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return dashboard.Run()
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
