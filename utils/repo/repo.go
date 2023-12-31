package repo

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/eng618/eng/utils/log"
)

// IsDirty verifies the supplied repository path is in a clean state.
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

// PullLatestCode pulls the latest code for the supplied repository.
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

// EnsureOnMaster verifies repo is on the master branch, and checks it out if it is not.
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
