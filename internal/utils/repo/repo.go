package repo

import (
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/eng618/eng/internal/utils/log"
)

// IsDirty checks if the repository at the given path has uncommitted changes.
// It takes the repository path `repoPath` as input and returns a boolean indicating
// whether the repository is dirty and an error if any occurs.
func IsDirty(repoPath string) (bool, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return false, err
	}

	w, err := r.Worktree()
	if err != nil {
		return false, err
	}

	status, err := w.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}

// PullLatestCode pulls the latest changes from the current branch of the repository.
// It takes the repository path `repoPath` as input and returns an error if the operation fails.
// The function automatically detects the current branch and pulls from the corresponding remote.
func PullLatestCode(repoPath string) error {
	// Get current branch
	currentBranch, err := GetCurrentBranch(repoPath)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	refName := plumbing.NewBranchReferenceName(currentBranch)
	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: refName,
		Progress:      log.Writer(),
	})

	return err
}

// EnsureOnDefaultBranch ensures that the repository is on the default branch (main or master).
// It dynamically detects the default branch and switches to it if necessary.
// It takes the repository path `repoPath` as input and returns an error if the operation fails.
func EnsureOnDefaultBranch(repoPath string) error {
	// Get the main branch name for this repository
	mainBranch, err := GetMainBranch(repoPath)
	if err != nil {
		return err
	}

	// Get current branch
	currentBranch, err := GetCurrentBranch(repoPath)
	if err != nil {
		return err
	}

	// If we're already on the main branch, no need to switch
	if currentBranch == mainBranch {
		log.Success("Already on default branch: %s", mainBranch)
		return nil
	}

	// Switch to the main branch
	log.Warn("Currently on %s, attempting to switch to default branch: %s", currentBranch, mainBranch)

	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	mainRef := plumbing.NewBranchReferenceName(mainBranch)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: mainRef,
		Force:  true, // Force checkout even if the working tree is dirty
	})
	if err != nil {
		return err
	}

	log.Success("Switched to default branch: %s", mainBranch)
	return nil
}

// FetchBareRepo fetches updates for a bare repository.
// It takes the repository path `repoPath` and work tree `workTree` as inputs and
// returns an error if the operation fails.
func FetchBareRepo(repoPath, workTree string) error {
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+workTree, "fetch", "--all", "--prune")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	err := cmd.Run()
	if err != nil {
		log.Error("FetchBareRepo failed: %v", err)
		return err
	}

	return nil
}

// PullRebaseBareRepo performs a pull with rebase operation for a bare repository.
// It takes the repository path `repoPath` and work tree `workTree` as inputs and
// returns an error if the operation fails.
func PullRebaseBareRepo(repoPath, workTree string) error {
	// Get the current branch for the bare repository
	cmd := exec.Command("git", "--git-dir="+repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		log.Error("Failed to get current branch: %v", err)
		return err
	}
	currentBranch := strings.TrimSpace(string(output))

	if currentBranch == "" {
		log.Error("No current branch found in bare repository")
		return err
	}

	log.Info("Pulling branch: %s", currentBranch)

	// Pull with explicit remote and branch
	cmd = exec.Command(
		"git",
		"--git-dir="+repoPath,
		"--work-tree="+workTree,
		"pull",
		"--rebase",
		"--autostash",
		"--progress",
		"origin",
		currentBranch,
	)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	err = cmd.Run()
	if err != nil {
		log.Error("PullRebaseBareRepo failed: %v", err)
		return err
	}

	return nil
}

// GetMainBranch returns the main branch name for the repository (main or master).
// It checks for both main and master branches and returns the one that exists.
func GetMainBranch(repoPath string) (string, error) {
	// First check if main branch exists
	if branchExists(repoPath, "main") {
		return "main", nil
	}

	// Then check if master branch exists
	if branchExists(repoPath, "master") {
		return "master", nil
	}

	// If neither exists, try to get the default branch from remote
	defaultBranch, err := getRemoteDefaultBranch(repoPath)
	if err == nil && defaultBranch != "" {
		return defaultBranch, nil
	}

	// Fall back to main as default
	log.Warn("Could not determine main branch for %s, defaulting to 'main'", repoPath)
	return "main", nil
}

// GetDevelopBranch returns the development branch name for the repository (develop, dev, or development).
// It checks for common development branch names and returns the one that exists.
func GetDevelopBranch(repoPath string) (string, error) {
	developBranches := []string{"develop", "dev", "development"}

	for _, branch := range developBranches {
		if branchExists(repoPath, branch) {
			return branch, nil
		}
	}

	// If no development branch is found, return empty string
	return "", nil
}

// GetCurrentBranch returns the current branch name for the repository.
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// branchExists checks if a branch exists in the repository.
func branchExists(repoPath, branchName string) bool {
	cmd := exec.Command("git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	err := cmd.Run()
	return err == nil
}

// getRemoteDefaultBranch tries to get the default branch from the remote.
func getRemoteDefaultBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse the output to get just the branch name
	ref := strings.TrimSpace(string(output))
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return "", nil
}

// CheckoutBareRepo performs a git checkout operation for a bare repository.
// It takes the repository path `repoPath` and work tree `workTree` as inputs.
// If force is true, it will discard any local changes and force the checkout.
// If all is true, it will checkout all files from the index/HEAD.
// Returns an error if the operation fails.
func CheckoutBareRepo(repoPath, workTree string, force, all bool) error {
	args := []string{"--git-dir=" + repoPath, "--work-tree=" + workTree, "checkout"}

	if force {
		args = append(args, "--force")
	}

	if all {
		args = append(args, ".")
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()

	err := cmd.Run()
	if err != nil {
		log.Error("CheckoutBareRepo failed: %v", err)
		return err
	}

	return nil
}
