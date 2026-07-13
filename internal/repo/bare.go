package repo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/eng618/eng/internal/log"
)

// BareClone clones a git repository as a bare repository.
func BareClone(ctx context.Context, repoURL, branch, bareRepoPath string, auth *ssh.PublicKeys) error {
	_, err := git.PlainCloneContext(ctx, bareRepoPath, true, &git.CloneOptions{
		URL:           repoURL,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Progress:      log.Writer(),
	})
	if err != nil {
		return wrapCloneError(repoURL, err)
	}
	return nil
}

// CheckoutWorktree checks out the tree from a bare repository to a worktree directory.
func CheckoutWorktree(ctx context.Context, bareRepoPath, worktreeDir string) error {
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("failed to get tree: %w", err)
	}

	err = tree.Files().ForEach(func(f *object.File) error {
		filePath := filepath.Join(worktreeDir, f.Name)

		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", f.Name, err)
		}

		contents, err := f.Contents()
		if err != nil {
			return fmt.Errorf("failed to get contents of %s: %w", f.Name, err)
		}

		if err := os.WriteFile(filePath, []byte(contents), os.FileMode(f.Mode)); err != nil {
			return fmt.Errorf("failed to write %s: %w", f.Name, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to checkout files: %w", err)
	}

	return nil
}

// InitSubmodules initializes and updates git submodules for a bare repository.
func InitSubmodules(ctx context.Context, bareRepoPath, worktreeDir string, auth *ssh.PublicKeys) error {
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		// Fall back to git command line
		cmd := exec.CommandContext(ctx, // #nosec G204
			"git",
			"--git-dir="+bareRepoPath,
			"--work-tree="+worktreeDir,
			"submodule",
			"update",
			"--init",
			"--recursive",
		)
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		return cmd.Run()
	}

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
	return nil
}

// ConfigureBareRepo sets the bare repository to not show untracked files.
func ConfigureBareRepo(ctx context.Context, bareRepoPath string) error {
	repo, err := git.PlainOpen(bareRepoPath)
	if err != nil {
		return err
	}
	cfg, err := repo.Config()
	if err != nil {
		return err
	}
	cfg.Raw.SetOption("status", "", "showUntrackedFiles", "no")
	return repo.Storer.SetConfig(cfg)
}

func wrapCloneError(repoURL string, err error) error {
	errMsg := strings.ToLower(err.Error())
	isSSH := strings.HasPrefix(repoURL, "git@") || strings.HasPrefix(repoURL, "ssh://")
	if isSSH &&
		(strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "publickey")) {
		return fmt.Errorf(
			"failed to clone repository via SSH: %w. Verify ~/.ssh/github works for GitHub or run `eng system setup ssh`",
			err,
		)
	}
	return fmt.Errorf("failed to clone repository: %w", err)
}
