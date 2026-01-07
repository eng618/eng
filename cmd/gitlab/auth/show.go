package auth

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "github.com/eng618/eng/utils"
    "github.com/eng618/eng/utils/log"
)

// showCmd prints effective GitLab auth/config details without exposing secrets
var showCmd = &cobra.Command{
    Use:   "show",
    Short: "Show GitLab auth and defaults (no secrets)",
    RunE: func(cmd *cobra.Command, args []string) error {
        host := viper.GetString("gitlab.host")
        project := viper.GetString("gitlab.project")
        tokenItem := viper.GetString("gitlab.tokenItem")

        // Determine token source availability
        tokenSource := ""
        if os.Getenv("GITLAB_TOKEN") != "" {
            tokenSource = "env:GITLAB_TOKEN"
        } else if tokenItem != "" {
            // Try to locate item without printing its content
            if _, err := utils.GetBitwardenItem(tokenItem); err == nil {
                tokenSource = fmt.Sprintf("bitwarden:%s", tokenItem)
            } else {
                tokenSource = fmt.Sprintf("bitwarden:%s (not found)", tokenItem)
            }
        } else if viper.GetString("gitlab.token") != "" {
            tokenSource = "config:gitlab.token (discouraged)"
        } else {
            tokenSource = "none"
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
