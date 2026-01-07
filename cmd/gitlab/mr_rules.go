package gitlab

import "github.com/spf13/cobra"

// mrRulesCmd is the parent command for managing MR rules
var mrRulesCmd = &cobra.Command{
    Use:   "mr-rules",
    Short: "Manage GitLab merge request rules",
}

func init() {
    GitLabCmd.AddCommand(mrRulesCmd)
}
