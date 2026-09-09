package gitlab

import (
	"github.com/spf13/cobra"

	gitlabauth "github.com/eng618/eng/cmd/gitlab/auth"
	"github.com/eng618/eng/internal/execx"
)

var execCommand = execx.Command

// GitLabCmd represents the parent command for GitLab-related operations.
var GitLabCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "Interact with GitLab via glab",
	Long:  "Commands that integrate with GitLab using the glab CLI.",
}

func init() {
	GitLabCmd.AddCommand(gitlabauth.AuthCmd)
}
