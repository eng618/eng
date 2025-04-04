package repo

import (
	"os"
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
		Progress:      os.Stdout,
	})

	return err
}

// EnsureOnMaster ensures that the repository is on the master branch.
// It takes the repository path `repoPath` as input and returns an error if the
// repository is not on the master branch or if the operation fails.
func EnsureOnMaster(repoPath string) error {
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

	if h.Name().Short() != "master" {
		log.Warn("head is currently at %s, attempting to switch to master", h.Name().Short())
		err = w.Checkout(&git.CheckoutOptions{
			Force: true,
		})
		if err != nil {
			return err
		}
	}

	log.Success("you are now on master")

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
