package dotfiles

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CopyChangesCmd defines the cobra command for copying modified dotfiles to the local git repository.
var CopyChangesCmd = &cobra.Command{
	Use:   "copy-changes",
	Short: "copy modified dotfiles to local git repo",
	Long:  `This command copies modified dotfiles from the worktree to the local git repository for committing.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Copying modified dotfiles")

		isVerbose := utils.IsVerbose(cmd)

		repoPath, worktreePath, err := getDotfilesConfig()
		if err != nil || repoPath == "" {
			log.Error("Dotfiles repository path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Repository path: %s", repoPath)
		log.Verbose(isVerbose, "Worktree path:   %s", worktreePath)

		devPath := os.ExpandEnv(viper.GetString("git.dev_path"))
		if devPath == "" {
			log.Error("Development folder path is not set in configuration")
			return
		}
		log.Verbose(isVerbose, "Development path: %s", devPath)

		engCfgPath := filepath.Join(devPath, "eng-cfg")
		log.Verbose(isVerbose, "eng-cfg path: %s", engCfgPath)

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
			dest := filepath.Join(engCfgPath, file)

			// Ensure destination directory exists
			destDir := filepath.Dir(dest)
			if err := os.MkdirAll(destDir, 0755); err != nil {
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
		prompt := &survey.Confirm{
			Message: "Do you want to reset the local copies in the worktree?",
		}
		prompt.Default = true
		err = survey.AskOne(prompt, &resetConfirm)
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

// getModifiedFilesFunc is injectable for tests
var getModifiedFilesFunc = func(repoPath, worktreePath string) ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+worktreePath, "status", "--porcelain")
	cmd.Stdout = &buf
	cmd.Stderr = log.ErrorWriter()
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " M ") || strings.HasPrefix(line, "M ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				files = append(files, parts[1])
			}
		}
	}

	return files, scanner.Err()
}

// resetFile runs git checkout -- file
func resetFile(repoPath, worktreePath, file string) error {
	cmd := exec.Command("git", "--git-dir="+repoPath, "--work-tree="+worktreePath, "checkout", "--", file)
	cmd.Dir = worktreePath // Run from worktree directory
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	return cmd.Run()
}

// copyFile copies a file from srcPath to destPath
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
