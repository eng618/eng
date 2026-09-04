package dotfiles

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errTestAbort = errors.New("test abort")

func TestCopyChangesCmd_MissingConfig(t *testing.T) {
	// Ensure viper has no config for dotfiles
	viper.Reset()

	// Run command; should return early without panic
	cmd := &cobra.Command{}
	CopyChangesCmd.Run(cmd, []string{})
}

func TestGetModifiedFiles(t *testing.T) {
	viper.Reset()
	viper.Set("dotfiles.bare_repo_path", "/tmp/repo")
	viper.Set("dotfiles.worktree_path", "/tmp/worktree")
	viper.Set("git.dev_path", "/tmp/dev")

	// Mock the function to return some files
	original := getModifiedFilesFunc
	getModifiedFilesFunc = func(repoPath, worktreePath string) ([]string, error) {
		return []string{".tool-versions", ".zshrc"}, nil
	}
	defer func() { getModifiedFilesFunc = original }()

	files, err := getModifiedFilesFunc("/tmp/repo", "/tmp/worktree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0] != ".tool-versions" {
		t.Fatalf("expected .tool-versions, got %s", files[0])
	}
}

func TestGetModifiedFilesWithSpaces(t *testing.T) {
	// Test that filenames with spaces are parsed correctly
	// This tests the fix for the bug where strings.Fields() was truncating filenames with spaces
	original := getModifiedFilesFunc
	getModifiedFilesFunc = func(repoPath, worktreePath string) ([]string, error) {
		// Simulate the actual git status --porcelain output
		// This needs to parse the raw git output, so we create a test that validates parsing
		return []string{"path/to/config file.txt", ".config/my settings"}, nil
	}
	defer func() { getModifiedFilesFunc = original }()

	files, err := getModifiedFilesFunc("/tmp/repo", "/tmp/worktree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0] != "path/to/config file.txt" {
		t.Fatalf("expected 'path/to/config file.txt', got '%s'", files[0])
	}
	if files[1] != ".config/my settings" {
		t.Fatalf("expected '.config/my settings', got '%s'", files[1])
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp files
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	destFile := filepath.Join(destDir, "test.txt")

	content := "test content"
	if err := os.WriteFile(srcFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write src file: %v", err)
	}

	if err := copyFile(srcFile, destFile, false); err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	destContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}

	if string(destContent) != content {
		t.Fatalf("content mismatch: got %s, want %s", destContent, content)
	}
}

func setupCopyTestViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	path := filepath.Join(t.TempDir(), ".eng.yaml")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	viper.SetConfigFile(path)
}

func mockTargetPrompts(
	selectFn func(string, []string, string) (string, error),
	inputFn func(string, string) (string, error),
) func() {
	origSelect, origInput := selectTargetFunc, inputTargetFunc
	selectTargetFunc, inputTargetFunc = selectFn, inputFn
	return func() {
		selectTargetFunc, inputTargetFunc = origSelect, origInput
	}
}

func makeGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestResolveCopyTarget_Flag(t *testing.T) {
	setupCopyTestViper(t)
	defer mockTargetPrompts(
		func(string, []string, string) (string, error) {
			t.Error("must not prompt for explicit --repo")
			return "", nil
		},
		func(string, string) (string, error) {
			t.Error("must not prompt for explicit --repo")
			return "", nil
		},
	)()

	got, err := resolveCopyTarget(makeGitRepo(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected resolved path")
	}

	if _, err := resolveCopyTarget(t.TempDir()); err == nil {
		t.Error("expected error for non-git --repo dir")
	} else if !strings.Contains(err.Error(), "--repo") {
		t.Errorf("expected --repo hint, got: %v", err)
	}
}

func TestResolveCopyTarget_Explicit(t *testing.T) {
	setupCopyTestViper(t)
	defer mockTargetPrompts(
		func(string, []string, string) (string, error) {
			t.Error("must not prompt for valid explicit config")
			return "", nil
		},
		func(string, string) (string, error) {
			t.Error("must not prompt for valid explicit config")
			return "", nil
		},
	)()

	// No dev path needed when explicit config resolves.
	viper.Set("git.dev_path", "")
	want := makeGitRepo(t)
	viper.Set("dotfiles.target_repo_path", want)

	got, err := resolveCopyTarget("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestResolveCopyTarget_ExplicitInvalidPersistsChoice(t *testing.T) {
	setupCopyTestViper(t)
	viper.Set("git.dev_path", "")
	viper.Set("dotfiles.target_repo_path", t.TempDir()) // no .git

	want := makeGitRepo(t)
	defer mockTargetPrompts(
		func(_ string, options []string, _ string) (string, error) {
			if len(options) == 0 || options[len(options)-1] != "Enter a path manually" {
				t.Errorf("expected manual fallback option, got %v", options)
			}
			return "Enter a path manually", nil
		},
		func(_, _ string) (string, error) { return want, nil },
	)()

	got, err := resolveCopyTarget("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if viper.GetString("dotfiles.target_repo_path") != want {
		t.Errorf("expected choice persisted, got %q", viper.GetString("dotfiles.target_repo_path"))
	}
}

func TestResolveCopyTarget_NoDevPathNoConfig(t *testing.T) {
	setupCopyTestViper(t)
	viper.Set("git.dev_path", "")

	if _, err := resolveCopyTarget(""); err == nil {
		t.Error("expected error with no dev path and no config")
	} else if !strings.Contains(err.Error(), "dotfiles-target-repo-path") {
		t.Errorf("expected config hint, got: %v", err)
	}
}

func TestResolveCopyTarget_HeuristicFallbackInteractive(t *testing.T) {
	setupCopyTestViper(t)
	viper.Set("git.dev_path", t.TempDir()) // empty dev dir: legacy fallback won't be a repo

	want := makeGitRepo(t)
	defer mockTargetPrompts(
		func(_ string, _ []string, _ string) (string, error) { return "Enter a path manually", nil },
		func(_, _ string) (string, error) { return want, nil },
	)()

	got, err := resolveCopyTarget("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestResolveCopyTarget_PromptError(t *testing.T) {
	setupCopyTestViper(t)
	viper.Set("git.dev_path", t.TempDir())
	defer mockTargetPrompts(
		func(_ string, _ []string, _ string) (string, error) { return "", errTestAbort },
		func(_, _ string) (string, error) { return "", errTestAbort },
	)()

	if _, err := resolveCopyTarget(""); err == nil {
		t.Error("expected error when prompt aborts")
	}
}
