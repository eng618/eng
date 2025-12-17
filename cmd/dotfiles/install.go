package dotfiles

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/eng618/eng/cmd/system"
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dotfiles from a bare git repository",
	Long: `Install dotfiles from a bare git repository. This command will:
  - Check and install prerequisites (Homebrew, Git, Bash, GitHub CLI, SSH keys)
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
  - Configure git to hide untracked files`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := installDotfiles(); err != nil {
			log.Fatal("Dotfiles installation failed: %v", err)
		}
	},
}

// installDotfiles orchestrates the complete dotfiles installation workflow.
func installDotfiles() error {
	log.Start("Starting dotfiles installation")

	// Step 1: Check prerequisites
	if err := system.EnsurePrerequisites(); err != nil {
		return err
	}

	// Step 2: Get configuration values
	repoURL := config.RepoURL()
	branch := config.Branch()
	bareRepoPath := config.BareRepoPath()
	bareRepoPath = os.ExpandEnv(bareRepoPath)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
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
		case "update":
			if err := updateBareRepo(bareRepoPath); err != nil {
				return err
			}
		case "fresh":
			if err := freshInstall(bareRepoPath, repoURL, branch); err != nil {
				return err
			}
		}
	} else {
		// Step 4: Clone bare repository
		if err := cloneBareRepo(repoURL, branch, bareRepoPath); err != nil {
			return err
		}
	}

	// Step 5: Backup conflicting files
	backupPath, hasConflicts, err := backupConflicts(bareRepoPath, homeDir)
	if err != nil {
		return err
	}

	// Step 6: Checkout files
	if err := checkoutFiles(bareRepoPath, homeDir); err != nil {
		if hasConflicts {
			log.Error("Checkout failed. Conflicting files have been backed up to: %s", backupPath)
			log.Message("You can restore files from the backup if needed")
		}
		return err
	}

	// Step 7: Initialize submodules
	if err := initSubmodules(bareRepoPath, homeDir); err != nil {
		log.Warn("Failed to initialize submodules: %v", err)
		log.Message("You can manually initialize them later with: git --git-dir=%s --work-tree=%s submodule update --init --recursive", bareRepoPath, homeDir)
	}

	// Step 8: Configure git
	if err := configureGit(bareRepoPath, homeDir); err != nil {
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

	// Extract the action keyword
	parts := strings.SplitN(action, " - ", 2)
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "skip", nil
}

// updateBareRepo fetches and pulls the latest changes in the bare repository.
func updateBareRepo(bareRepoPath string) error {
	log.Start("Updating existing repository")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	// Fetch updates
	fetchCmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "fetch", "origin")
	fetchCmd.Stdout = log.Writer()
	fetchCmd.Stderr = log.ErrorWriter()

	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch updates: %w", err)
	}

	// Pull with rebase
	pullCmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "pull", "--rebase")
	pullCmd.Stdout = log.Writer()
	pullCmd.Stderr = log.ErrorWriter()

	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull updates: %w", err)
	}

	log.Success("Repository updated successfully")
	return nil
}

// freshInstall deletes the existing repository and clones a fresh copy.
func freshInstall(bareRepoPath, repoURL, branch string) error {
	log.Start("Removing existing repository")

	if err := os.RemoveAll(bareRepoPath); err != nil {
		return fmt.Errorf("failed to remove existing repository: %w", err)
	}

	log.Success("Existing repository removed")

	return cloneBareRepo(repoURL, branch, bareRepoPath)
}

// cloneBareRepo clones the dotfiles repository as a bare repository.
func cloneBareRepo(repoURL, branch, bareRepoPath string) error {
	log.Start("Cloning repository from: %s (branch: %s)", repoURL, branch)

	cmd := exec.Command("gh", "repo", "clone", repoURL, bareRepoPath, "--", "--bare", "--branch", branch)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Success("Repository cloned successfully")
	return nil
}

// backupConflicts identifies and backs up files that would conflict with the checkout.
// Returns the backup path, whether conflicts were found, and any error.
func backupConflicts(bareRepoPath, homeDir string) (string, bool, error) {
	log.Start("Checking for conflicting files")

	// Create timestamped backup directory
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(homeDir, fmt.Sprintf(".config-backup-%s", timestamp))

	// Get list of tracked files
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "ls-tree", "-r", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", false, fmt.Errorf("failed to list tracked files: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	conflictCount := 0

	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file == "" {
			continue
		}

		filePath := filepath.Join(homeDir, file)

		// Check if file exists (regular file or symlink)
		if _, err := os.Lstat(filePath); err == nil {
			// File exists - back it up
			if conflictCount == 0 {
				// Create backup directory on first conflict
				if err := os.MkdirAll(backupPath, 0755); err != nil {
					return "", false, fmt.Errorf("failed to create backup directory: %w", err)
				}
			}

			backupFilePath := filepath.Join(backupPath, file)

			// Create parent directories in backup
			if err := os.MkdirAll(filepath.Dir(backupFilePath), 0755); err != nil {
				log.Warn("Failed to create backup directory for %s: %v", file, err)
				continue
			}

			// Move file to backup
			if err := os.Rename(filePath, backupFilePath); err != nil {
				log.Warn("Failed to backup %s: %v", file, err)
				continue
			}

			log.Message("Backed up: %s", file)
			conflictCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("error reading tracked files: %w", err)
	}

	if conflictCount > 0 {
		log.Success("Backed up %d conflicting file(s) to: %s", conflictCount, backupPath)
		return backupPath, true, nil
	}

	log.Success("No conflicting files found")
	return "", false, nil
}

// checkoutFiles checks out the dotfiles from the bare repository to the home directory.
func checkoutFiles(bareRepoPath, homeDir string) error {
	log.Start("Checking out files to home directory")

	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "checkout")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout files: %w", err)
	}

	log.Success("Files checked out successfully")
	return nil
}

// initSubmodules initializes and updates git submodules.
func initSubmodules(bareRepoPath, homeDir string) error {
	log.Start("Initializing git submodules")

	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "submodule", "update", "--init", "--recursive")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		return err
	}

	log.Success("Git submodules initialized successfully")
	return nil
}

// configureGit sets git configuration to hide untracked files.
func configureGit(bareRepoPath, homeDir string) error {
	log.Start("Configuring git settings")

	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "config", "status.showUntrackedFiles", "no")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	if err := cmd.Run(); err != nil {
		return err
	}

	log.Success("Git configured successfully")
	return nil
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
