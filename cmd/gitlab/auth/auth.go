package auth

import "github.com/spf13/cobra"

// AuthCmd is the parent for gitlab auth commands
var AuthCmd = &cobra.Command{
    Use:   "auth",
    Short: "Manage GitLab authentication for eng",
}
