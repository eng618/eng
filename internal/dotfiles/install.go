package dotfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/eng618/eng/cmd/system"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/repo"
)

// Install orchestrates the complete dotfiles installation workflow.
func Install(ctx context.Context, verbose bool) error {
	log.Start("Starting dotfiles installation")

	// Step 1: Get configuration values first so we can make context-aware setup decisions.
	repoURL, branch, bareRepoPath, worktreePath, err := config.VerifyDotfilesConfig()
	if err != nil {
		return err
	}

	// Step 2: Check prerequisites
	if err := system.EnsurePrerequisites(verbose); err != nil {
		return err
	}

	// Step 3: Handle existing bare repository
	if _, err := os.Stat(bareRepoPath); err == nil {
		action, err := handleExistingRepo(bareRepoPath)
		if err != nil {
			return err
		}

		switch action {
		case "skip":
			log.Message("Skipping repository clone, using existing repository")
			return nil
		case "update":
			if err := ensureSSHIfRequired(repoURL, verbose); err != nil {
				return err
			}
			if err := updateBareRepoWorktree(ctx, bareRepoPath, worktreePath, verbose); err != nil {
				return err
			}
			return nil
		case "fresh":
			if err := ensureSSHIfRequired(repoURL, verbose); err != nil {
				return err
			}
			if err := freshInstall(ctx, bareRepoPath, repoURL, branch); err != nil {
				return err
			}
			// Fall through to backup and checkout for fresh install
		}
	} else {
		// Step 4: Clone bare repository
		if err := ensureSSHIfRequired(repoURL, verbose); err != nil {
			return err
		}
		if err := cloneBareRepo(ctx, repoURL, branch, bareRepoPath); err != nil {
			return err
		}
	}

	// Step 5: Backup conflicting files
	backupPath, hasConflicts, err := backupConflicts(bareRepoPath, worktreePath)
	if err != nil {
		return err
	}

	// Step 6: Checkout files
	if err := repo.CheckoutWorktree(ctx, bareRepoPath, worktreePath); err != nil {
		if hasConflicts {
			log.Error("Checkout failed. Conflicting files have been backed up to: %s", backupPath)
			log.Message("You can restore files from the backup if needed")
		}
		return err
	}

	// Step 7: Initialize submodules
	auth, _ := getSSHAuth() // best effort
	if err := repo.InitSubmodules(ctx, bareRepoPath, worktreePath, auth); err != nil {
		log.Warn("Failed to initialize submodules: %v", err)
		log.Message(
			"You can manually initialize them later with: git --git-dir=%s --work-tree=%s submodule update --init --recursive",
			bareRepoPath,
			worktreePath,
		)
	}

	// Step 8: Configure git
	if err := repo.ConfigureBareRepo(ctx, bareRepoPath); err != nil {
		log.Warn("Failed to configure git: %v", err)
	}

	// Step 9: Print instructions
	printCompletionInstructions(bareRepoPath, hasConflicts, backupPath)

	log.Success("Dotfiles installation completed successfully")
	return nil
}

// handleExistingRepo prompts the user for what to do with an existing bare repository.
func handleExistingRepo(bareRepoPath string) (string, error) {
	log.Warn("Bare repository already exists at: %s", bareRepoPath)

	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"skip - Use existing repository without changes",
			"update - Fetch and pull latest changes",
			"fresh - Delete and re-clone repository",
		},
		Default: "skip - Use existing repository without changes",
	}

	if err := survey.AskOne(prompt, &action); err != nil {
		return "", err
	}

	parts := strings.SplitN(action, " - ", 2)
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "skip", nil
}

// updateBareRepoWorktree fetches and pulls the latest changes in the bare repository to the worktree.
func updateBareRepoWorktree(ctx context.Context, bareRepoPath, homeDir string, isVerbose bool) error {
	log.Start("Updating existing repository and worktree")

	if err := SyncRepo(ctx, bareRepoPath, homeDir, isVerbose); err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	log.Success("Repository and worktree updated successfully")
	return nil
}

// freshInstall deletes the existing repository and clones a fresh copy.
func freshInstall(ctx context.Context, bareRepoPath, repoURL, branch string) error {
	log.Start("Removing existing repository")

	if err := os.RemoveAll(bareRepoPath); err != nil {
		return fmt.Errorf("failed to remove existing repository: %w", err)
	}

	log.Success("Existing repository removed")

	return cloneBareRepo(ctx, repoURL, branch, bareRepoPath)
}

// cloneBareRepo clones the dotfiles repository as a bare repository.
func cloneBareRepo(ctx context.Context, repoURL, branch, bareRepoPath string) error {
	log.Start("Cloning repository from: %s (branch: %s)", repoURL, branch)

	var auth *ssh.PublicKeys
	var err error
	if isSSHRepoURL(repoURL) {
		auth, err = getSSHAuth()
		if err != nil {
			return fmt.Errorf("failed to get SSH auth from ~/.ssh/github: %w", err)
		}
	}

	if err := repo.BareClone(ctx, repoURL, branch, bareRepoPath, auth); err != nil {
		return err
	}

	log.Success("Repository cloned successfully")
	return nil
}

// backupConflicts identifies and backs up files that would conflict with the checkout.
func backupConflicts(bareRepoPath, homeDir string) (string, bool, error) {
	log.Start("Checking for conflicting files")

	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(homeDir, fmt.Sprintf(".config-backup-%s", timestamp))

	r, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := r.Head()
	if err != nil {
		return "", false, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return "", false, fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", false, fmt.Errorf("failed to get tree: %w", err)
	}

	conflictCount := 0

	err = tree.Files().ForEach(func(f *object.File) error {
		file := f.Name
		filePath := filepath.Join(homeDir, file)

		if _, err := os.Lstat(filePath); err == nil {
			if conflictCount == 0 {
				if err := os.MkdirAll(backupPath, 0o755); err != nil {
					return fmt.Errorf("failed to create backup directory: %w", err)
				}
			}

			backupFilePath := filepath.Join(backupPath, file)
			if err := os.MkdirAll(filepath.Dir(backupFilePath), 0o755); err != nil {
				log.Warn("Failed to create backup directory for %s: %v", file, err)
				return nil
			}

			if err := os.Rename(filePath, backupFilePath); err != nil {
				log.Warn("Failed to backup %s: %v", file, err)
				return nil
			}

			log.Message("Backed up: %s", file)
			conflictCount++
		}
		return nil
	})
	if err != nil {
		return "", false, fmt.Errorf("error reading tracked files: %w", err)
	}

	if conflictCount > 0 {
		log.Success("Backed up %d conflicting file(s) to: %s", conflictCount, backupPath)
		return backupPath, true, nil
	}

	log.Success("No conflicting files found")
	return "", false, nil
}

// printCompletionInstructions displays instructions for using the cfg alias and other important information.
func printCompletionInstructions(bareRepoPath string, hasConflicts bool, backupPath string) {
	homeDir, _ := os.UserHomeDir()

	log.Message("")
	log.Message("-----------------------------------------------------")
	log.Message(" Installation Complete!")
	log.Message("-----------------------------------------------------")
	log.Message("")

	if hasConflicts {
		log.Message(" BACKUP INFORMATION:")
		log.Message("   Conflicting files were backed up to:")
		log.Message("   %s", backupPath)
		log.Message("")
	}

	log.Message(" MANAGING YOUR DOTFILES:")
	log.Message("")
	log.Message("   Your dotfiles have been checked out to your home directory.")
	log.Message("   You can manage them using the 'cfg' alias (if configured in your")
	log.Message("   shell dotfiles).")
	log.Message("")
	log.Message("   If the alias is not yet available, reload your shell configuration:")
	log.Message("     source ~/.zshrc    (for zsh)")
	log.Message("     source ~/.bashrc   (for bash)")
	log.Message("")
	log.Message("   Common commands:")
	log.Message("     cfg status          # See changes")
	log.Message("     cfg add ~/.vimrc    # Stage a file")
	log.Message("     cfg commit -m 'Msg' # Commit changes")
	log.Message("     cfg push            # Push changes to remote")
	log.Message("     cfg pull            # Pull changes from remote")
	log.Message("")
	log.Message("   If you need to manually set up the alias, add this to your shell config:")
	log.Message("     alias cfg='git --git-dir=%s --work-tree=%s'", bareRepoPath, homeDir)
	log.Message("")
	log.Message("-----------------------------------------------------")
}

// getSSHAuth returns SSH authentication using the github SSH key.
func getSSHAuth() (*ssh.PublicKeys, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	sshKeyPath := filepath.Join(homeDir, ".ssh", "github")
	auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from %s: %w", sshKeyPath, err)
	}

	return auth, nil
}

func ensureSSHIfRequired(repoURL string, verbose bool) error {
	if !isSSHRepoURL(repoURL) {
		return nil
	}

	if err := system.SetupSSHForGitHub(verbose); err != nil {
		return fmt.Errorf("unable to setup GitHub SSH access: %w", err)
	}

	return nil
}

func isSSHRepoURL(repoURL string) bool {
	return strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://")
}
