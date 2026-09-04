package dotfiles

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// CopyChangesCmd defines the cobra command for copying modified dotfiles to the local git repository.
var CopyChangesCmd = &cobra.Command{
	Use:   "copy-changes",
	Short: "copy modified dotfiles to local git repo",
	Long: `This command copies modified dotfiles from the worktree to the local git repository for committing.

The destination resolves as: --repo flag, explicit config
('eng config dotfiles-target-repo-path'), then dev-folder heuristics.
When nothing resolves to a git repository, it offers to locate one
interactively and persists the choice.`,
	Example: `  eng dotfiles copy-changes
  eng dotfiles copy-changes --repo ~/Development/dotfiles`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Copying modified dotfiles")

		isVerbose := cmdutil.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		repoFlag, _ := cmd.Flags().GetString("repo")
		targetRepoPath, err := resolveCopyTarget(repoFlag)
		if err != nil {
			log.Error("%s", err)
			return
		}
		log.Verbose(isVerbose, "Target repository path: %s", targetRepoPath)

		// Get modified files
		modifiedFiles, err := getModifiedFilesFunc(repoPath, worktreePath)
		if err != nil {
			log.Error("Failed to get modified files: %s", err)
			return
		}

		if len(modifiedFiles) == 0 {
			log.Info("No modified files found")
			return
		}

		log.Info("Found %d modified files", len(modifiedFiles))

		// Copy files
		for _, file := range modifiedFiles {
			src := filepath.Join(worktreePath, file)
			dest := filepath.Join(targetRepoPath, file)

			// Ensure destination directory exists
			destDir := filepath.Dir(dest)
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				log.Error("Failed to create directory %s: %s", destDir, err)
				continue
			}

			if err := copyFile(src, dest, isVerbose); err != nil {
				log.Error("Failed to copy %s to %s: %s", src, dest, err)
				continue
			}

			log.Info("Copied %s to %s", file, dest)
		}

		log.Success("Copied modified dotfiles successfully")

		// Ask to reset
		var resetConfirm bool
		resetConfirm, err = ui.Confirm("Do you want to reset the local copies in the worktree?", true)
		if err != nil {
			log.Error("Failed to get user confirmation: %s", err)
			return
		}

		if resetConfirm {
			log.Start("Resetting local copies")
			for _, file := range modifiedFiles {
				if err := resetFile(repoPath, worktreePath, file); err != nil {
					log.Error("Failed to reset %s: %s", file, err)
					continue
				}
				log.Verbose(isVerbose, "Reset %s", file)
			}
			log.Success("Reset local copies successfully")
		}
	},
}

// resolveCopyTarget determines where modified files are copied.
// Precedence: --repo flag, explicit config, dev-folder heuristics. A
// resolved path must be a git repository; otherwise (or when nothing
// resolves) it falls back to interactive location, which persists the
// choice. devPath is only required for the heuristic search.
func resolveCopyTarget(repoFlag string) (string, error) {
	if flag := strings.TrimSpace(repoFlag); flag != "" {
		target := os.ExpandEnv(flag)
		if !config.IsGitRepo(target) {
			return "", fmt.Errorf(
				"not a git repository: %s — check --repo, or persist one with 'eng config dotfiles-target-repo-path <path>'",
				target,
			)
		}
		return target, nil
	}

	if explicit := config.ExplicitTargetRepoPath(); explicit != "" {
		if !config.IsGitRepo(explicit) {
			return locateTargetInteractively(explicit)
		}
		return explicit, nil
	}

	devPath := os.ExpandEnv(config.GetGitConfig().DevPath)
	if devPath == "" {
		return "", fmt.Errorf(
			"development folder path is not set and no target repository is configured — set one with 'eng config dotfiles-target-repo-path <path>' or 'eng config git-dev-path <path>'",
		)
	}
	target := getTargetRepoPathFunc(devPath)
	if !config.IsGitRepo(target) {
		return locateTargetInteractively(target)
	}
	return target, nil
}

// locateTargetInteractively helps the user point copy-changes at the right
// repository when automatic resolution fails, then persists the choice so
// the problem stays fixed. badTarget is the rejected path, if any.
func locateTargetInteractively(badTarget string) (string, error) {
	devPath := os.ExpandEnv(config.GetGitConfig().DevPath)
	var options []string
	if devPath != "" {
		options = append(options, config.TargetRepoCandidates(devPath)...)
	}
	options = append(options, "Enter a path manually")

	if badTarget != "" {
		log.Warn("Dotfiles target repo not found: %s", badTarget)
	}
	selected, err := selectTargetFunc(
		"Where should dotfile changes be copied?",
		options,
		options[0],
	)
	if err != nil {
		return "", fmt.Errorf(
			"no target repository selected — set one with 'eng config dotfiles-target-repo-path <path>': %w",
			err,
		)
	}
	if selected == "Enter a path manually" {
		selected, err = inputTargetFunc("Target repository path:", "")
		if err != nil {
			return "", fmt.Errorf(
				"no target repository entered — set one with 'eng config dotfiles-target-repo-path <path>': %w",
				err,
			)
		}
		selected = strings.TrimSpace(selected)
	}
	if !config.IsGitRepo(selected) {
		return "", fmt.Errorf(
			"not a git repository: %s — pick a directory containing .git",
			selected,
		)
	}
	if err := config.SetTargetRepoPath(selected); err != nil {
		return "", err
	}
	log.Info("Saved target repository path for future runs")
	return os.ExpandEnv(selected), nil
}

// getTargetRepoPathFunc is injectable for tests.
var getTargetRepoPathFunc = config.TargetRepoPath

// selectTargetFunc and inputTargetFunc are injectable for tests.
var (
	selectTargetFunc = ui.Select
	inputTargetFunc  = ui.Input
)

func init() {
	CopyChangesCmd.Flags().
		String("repo", "", "Destination git repository path (overrides config and heuristics)")
}

// getModifiedFilesFunc is injectable for tests.
var getModifiedFilesFunc = func(repoPath, worktreePath string) ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+worktreePath, "status", "--porcelain", "-z")
	cmd.Stdout = &buf
	cmd.Stderr = log.ErrorWriter()
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var files []string
	items := bytes.Split(buf.Bytes(), []byte{0})
	for _, item := range items {
		if len(item) == 0 {
			continue
		}
		line := string(item)
		// Git status format: XY PATH
		// X = staged status, Y = working tree status, followed by space and path starting at index 3
		if len(line) > 3 && (strings.HasPrefix(line, " M ") || strings.HasPrefix(line, "M ")) {
			files = append(files, line[3:])
		}
	}

	return files, nil
}

// resetFile runs git checkout -- file.
func resetFile(repoPath, worktreePath, file string) error {
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+worktreePath, "checkout", "--", file)
	cmd.Dir = worktreePath // Run from worktree directory
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	return cmd.Run()
}

// copyFile copies a file from srcPath to destPath.
func copyFile(srcPath, destPath string, isVerbose bool) error {
	log.Verbose(isVerbose, "Copying %s to %s", srcPath, destPath)

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := src.Close(); err != nil {
			log.Error("Failed to close source file %s: %s", srcPath, err)
		}
	}()

	srcInfo, err := src.Stat()
	if err != nil {
		return err
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() {
		if err := dest.Close(); err != nil {
			log.Error("Failed to close destination file %s: %s", destPath, err)
		}
	}()

	_, err = io.Copy(dest, src)
	return err
}
