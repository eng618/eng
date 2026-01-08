package dotfiles

import (
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
	"github.com/eng618/eng/internal/utils/repo"
)

// CheckoutCmd defines the cobra command for checking out files in the dotfiles repository.
// It can checkout all files or specific files, optionally with force to discard local changes.
var CheckoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "checkout files in your local bare repository",
	Long:  `This command checks out files from the dotfiles repository, optionally discarding local changes with the force flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Checking out dotfiles")

		isVerbose := utils.IsVerbose(cmd)
		all, _ := cmd.Flags().GetBool("all")
		force, _ := cmd.Flags().GetBool("force")

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		// Log what operation is being performed
		operation := "Checking out"
		if force {
			operation += " (force)"
		}
		if all {
			operation += " all files"
		} else {
			operation += " files"
		}
		log.Info(operation)

		// Use injectable function so tests can override and avoid executing git.
		err = checkoutRepo(repoPath, worktreePath, force, all)
		if err != nil {
			log.Error("Failed to checkout dotfiles: %s", err)
			return
		}

		log.Success("Dotfiles checked out successfully")
	},
}

func init() {
	CheckoutCmd.Flags().BoolP("all", "a", false, "checkout all files from the index/HEAD")
	CheckoutCmd.Flags().BoolP("force", "f", false, "force checkout, discarding any local changes")
}

// checkoutRepo is injectable for tests to avoid executing git.
var checkoutRepo = func(repoPath, worktreePath string, force, all bool) error {
	return repo.CheckoutBareRepo(repoPath, worktreePath, force, all)
}
