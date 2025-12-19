package dotfiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/system"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dotfiles from a bare git repository",
	Long: `Install dotfiles from a bare git repository. This command will:
  - Check and install prerequisites (Homebrew, Git, Bash, SSH keys)
  - Clone your dotfiles repository as a bare repository
  - Backup any conflicting files
  - Checkout dotfiles to your home directory
  - Initialize git submodules
  - Configure git to hide untracked files`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := installDotfiles(utils.IsVerbose(cmd)); err != nil {
			log.Fatal("Dotfiles installation failed: %v", err)
		}
	},
}

// installDotfiles orchestrates the complete dotfiles installation workflow.
func installDotfiles(verbose bool) error {
	log.Start("Starting dotfiles installation")

	// Step 1: Check prerequisites
	if err := system.EnsurePrerequisites(verbose); err != nil {
		return err
	}

	// Step 2: Get configuration values
	repoURL, branch, bareRepoPath, worktreePath := config.VerifyDotfilesConfig()

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
			if err := updateBareRepoWorktree(bareRepoPath, worktreePath, verbose); err != nil {
				return err
			}
			return nil
		case "fresh":
			if err := freshInstall(bareRepoPath, repoURL, branch); err != nil {
				return err
			}
			// Fall through to backup and checkout for fresh install
		}
	} else {
		// Step 4: Clone bare repository
		if err := cloneBareRepo(repoURL, branch, bareRepoPath); err != nil {
			return err
		}
	}

	// Step 5: Backup conflicting files
	backupPath, hasConflicts, err := backupConflicts(bareRepoPath, worktreePath)
	if err != nil {
		return err
	}

	// Step 6: Checkout files
	if err := checkoutFiles(bareRepoPath, worktreePath); err != nil {
		if hasConflicts {
			log.Error("Checkout failed. Conflicting files have been backed up to: %s", backupPath)
			log.Message("You can restore files from the backup if needed")
		}
		return err
	}

	// Step 7: Initialize submodules
	if err := initSubmodules(bareRepoPath, worktreePath); err != nil {
		log.Warn("Failed to initialize submodules: %v", err)
		log.Message("You can manually initialize them later with: git --git-dir=%s --work-tree=%s submodule update --init --recursive", bareRepoPath, worktreePath)
	}

	// Step 8: Configure git
	if err := configureGit(bareRepoPath, worktreePath); err != nil {
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

// updateBareRepoWorktree fetches and pulls the latest changes in the bare repository to the worktree.
// It uses the git command line tool because go-git has some limitations with bare repo worktrees.
func updateBareRepoWorktree(bareRepoPath, homeDir string, isVerbose bool) error {
	log.Start("Updating existing repository and worktree")

	if err := SyncRepo(bareRepoPath, homeDir, isVerbose); err != nil {
		return fmt.Errorf("failed to sync repository: %w", err)
	}

	log.Success("Repository and worktree updated successfully")
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

	// Get SSH auth
	auth, err := getSSHAuth()
	if err != nil {
		return fmt.Errorf("failed to get SSH auth: %w", err)
	}

	// Clone as bare repository
	_, err = git.PlainClone(bareRepoPath, true, &git.CloneOptions{
		URL:           repoURL,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Progress:      log.Writer(),
	})
	if err != nil {
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

	// Open the bare repository
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return "", false, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", false, fmt.Errorf("failed to get commit: %w", err)
	}

	// Get the tree
	tree, err := commit.Tree()
	if err != nil {
		return "", false, fmt.Errorf("failed to get tree: %w", err)
	}

	conflictCount := 0

	// Walk through all files in the tree
	err = tree.Files().ForEach(func(f *object.File) error {
		file := f.Name

		filePath := filepath.Join(homeDir, file)

		// Check if file exists (regular file or symlink)
		if _, err := os.Lstat(filePath); err == nil {
			// File exists - back it up
			if conflictCount == 0 {
				// Create backup directory on first conflict
				if err := os.MkdirAll(backupPath, 0o755); err != nil {
					return fmt.Errorf("failed to create backup directory: %w", err)
				}
			}

			backupFilePath := filepath.Join(backupPath, file)

			// Create parent directories in backup
			if err := os.MkdirAll(filepath.Dir(backupFilePath), 0o755); err != nil {
				log.Warn("Failed to create backup directory for %s: %v", file, err)
				return nil // Continue with next file
			}

			// Move file to backup
			if err := os.Rename(filePath, backupFilePath); err != nil {
				log.Warn("Failed to backup %s: %v", file, err)
				return nil // Continue with next file
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

// checkoutFiles checks out the dotfiles from the bare repository to the home directory.
func checkoutFiles(bareRepoPath, homeDir string) error {
	log.Start("Checking out files to home directory")

	// Open the bare repository
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	// Get the tree
	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get tree: %w", err)
	}

	// Walk through all files in the tree and checkout each one
	err = tree.Files().ForEach(func(f *object.File) error {
		filePath := filepath.Join(homeDir, f.Name)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", f.Name, err)
		}

		// Get file contents
		contents, err := f.Contents()
		if err != nil {
			return fmt.Errorf("failed to get contents of %s: %w", f.Name, err)
		}

		// Write file - convert go-git FileMode to os.FileMode
		if err := os.WriteFile(filePath, []byte(contents), os.FileMode(f.Mode)); err != nil {
			return fmt.Errorf("failed to write %s: %w", f.Name, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to checkout files: %w", err)
	}

	log.Success("Files checked out successfully")
	return nil
}

// initSubmodules initializes and updates git submodules.
func initSubmodules(bareRepoPath, homeDir string) error {
	log.Start("Initializing git submodules")

	// Open the bare repository
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return err
	}

	// Get worktree (we need to use a special approach for bare repos)
	w, err := repo.Worktree()
	if err != nil {
		// For bare repos, we can't get worktree directly
		// Fall back to using git command for submodules as they require worktree
		return initSubmodulesWithCommand(bareRepoPath, homeDir)
	}

	// Get SSH auth
	auth, err := getSSHAuth()
	if err != nil {
		return err
	}

	// Initialize submodules
	submodules, err := w.Submodules()
	if err != nil {
		return err
	}

	for _, submodule := range submodules {
		if err := submodule.Init(); err != nil {
			log.Warn("Failed to init submodule %s: %v", submodule.Config().Name, err)
			continue
		}

		if err := submodule.Update(&git.SubmoduleUpdateOptions{
			Init:              true,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Auth:              auth,
		}); err != nil {
			log.Warn("Failed to update submodule %s: %v", submodule.Config().Name, err)
			continue
		}
	}

	log.Success("Git submodules initialized successfully")
	return nil
}

// initSubmodulesWithCommand is a fallback for bare repositories where worktree isn't accessible.
func initSubmodulesWithCommand(bareRepoPath, homeDir string) error {
	// For bare repos, we need to use git command with work-tree
	// This is one case where shelling out is necessary due to go-git limitations with bare repos
	//nolint:gosec
	cmd := exec.Command("git", "--git-dir="+bareRepoPath, "--work-tree="+homeDir, "submodule", "update", "--init", "--recursive")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	return cmd.Run()
}

// configureGit sets git configuration to hide untracked files.
func configureGit(bareRepoPath, _ string) error {
	log.Start("Configuring git settings")

	// Open the bare repository
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return err
	}

	// Get repository config
	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	// Set status.showUntrackedFiles to no
	cfg.Raw.SetOption("status", "", "showUntrackedFiles", "no")

	// Save config
	if err := repo.Storer.SetConfig(cfg); err != nil {
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

// getSSHAuth returns SSH authentication using the github SSH key.
func getSSHAuth() (*ssh.PublicKeys, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	sshKeyPath := filepath.Join(homeDir, ".ssh", "github")

	// Read the private key
	auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from %s: %w", sshKeyPath, err)
	}

	return auth, nil
}
