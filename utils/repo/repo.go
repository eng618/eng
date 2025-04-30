package repo

import (
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/eng618/eng/utils/log"
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

// PullLatestCode pulls the latest changes from the specified branch of the repository.
// It takes the repository path `repoPath` and branch name `branchName` as inputs and
// returns an error if the operation fails.
func PullLatestCode(repoPath string, branchName string) error {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	refName := plumbing.NewBranchReferenceName(branchName)
	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: refName,
		Progress:      log.Writer(),
	})

	return err
}

// EnsureOnMain ensures that the repository is on the master branch.
// It takes the repository path `repoPath` as input and returns an error if the
// repository is not on the main branch or if the operation fails.
func EnsureOnMain(repoPath string) error {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	h, err := r.Head()
	if err != nil {
		return err
	}

	mainRef := plumbing.NewBranchReferenceName("main")

	if h.Name().Short() != "main" {
		log.Warn("head is currently at %s, attempting to switch to main", h.Name().Short())
		err = w.Checkout(&git.CheckoutOptions{
			Branch: mainRef,
			Force:  true, // Force checkout even if the working tree is dirty
		})
		if err != nil {
			return err
		}
		log.Success("Switched to main branch")
	} else {
		log.Success("Already on main branch")
	}

	return nil
}

// FetchBareRepo fetches updates for a bare repository.
// It takes the repository path `repoPath` and work tree `workTree` as inputs and
// returns an error if the operation fails.
func FetchBareRepo(repoPath string, workTree string) error {
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+workTree, "fetch", "--all", "--prune")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("FetchBareRepo output: %s", string(out)) // Log the output
		return err
	}
	log.Info("FetchBareRepo output: %s", string(out)) // Log the output

	return nil
}

// PullRebaseBareRepo performs a pull with rebase operation for a bare repository.
// It takes the repository path `repoPath` and work tree `workTree` as inputs and
// returns an error if the operation fails.
func PullRebaseBareRepo(repoPath string, workTree string) error {
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+workTree, "pull", "--rebase")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("PullRebaseBareRepo output: %s", string(out)) // Log the output
		return err
	}
	log.Info("PullRebaseBareRepo output: %s", string(out)) // Log the output

	return nil
}
