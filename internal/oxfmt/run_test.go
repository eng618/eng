package oxfmt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

// scriptRunner routes every command through the TestOxfmtHelperProcess
// test binary, which emulates oxfmt from scenario environment variables.
type scriptRunner struct {
	bins map[string]bool
}

func (r scriptRunner) CommandContext(_ context.Context, name string, args ...string) *exec.Cmd {
	argv := append([]string{"-test.run=TestOxfmtHelperProcess", "--", name}, args...)
	cmd := exec.Command(os.Args[0], argv...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func (r scriptRunner) LookPath(file string) (string, error) {
	if r.bins[file] {
		return "/test/bin/" + file, nil
	}
	return "", errors.New("not found: " + file)
}

// TestOxfmtHelperProcess emulates oxfmt/npx/bun/node for scriptRunner tests.
// It is a no-op under a normal test run (helper env var unset).
func TestOxfmtHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	sep := -1
	for i, a := range os.Args {
		if a == "--" {
			sep = i
			break
		}
	}
	if sep < 0 || sep+1 >= len(os.Args) {
		os.Exit(2)
	}
	rest := os.Args[sep+1:]
	name, args := rest[0], rest[1:]

	if name == "node" {
		version := os.Getenv("ENG_OXFMT_NODE_VERSION")
		if version == "" {
			version = "v26.8.2"
		}
		fmt.Fprintln(os.Stdout, version)
		os.Exit(0)
	}
	if slices.Contains(args, "--list-different") {
		if os.Getenv("ENG_OXFMT_SCENARIO") == "dirty" {
			fmt.Fprintln(os.Stdout, os.Getenv("ENG_OXFMT_DIRTY_FILE"))
			os.Exit(1)
		}
		os.Exit(0)
	}
	if slices.Contains(args, "--stdin-filepath") {
		if formatted := os.Getenv("ENG_OXFMT_FORMATTED"); formatted != "" {
			fmt.Fprint(os.Stdout, formatted)
		} else if in, err := io.ReadAll(os.Stdin); err == nil {
			fmt.Fprint(os.Stdout, string(in))
		}
		os.Exit(0)
	}
	if slices.Contains(args, "--write") {
		os.Exit(0)
	}
	os.Exit(2)
}

// captureLog redirects log writers into buffers for the duration of a test.
func captureLog(t *testing.T) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	var out, errOut bytes.Buffer
	log.SetWriters(&out, &errOut)
	t.Cleanup(log.ResetWriters)
	oldDisable := ui.DisableProgress
	ui.DisableProgress = true
	t.Cleanup(func() { ui.DisableProgress = oldDisable })
	return &out, &errOut
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.ts")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestRunCheckClean(t *testing.T) {
	out, _ := captureLog(t)
	path := writeTempFile(t, "const x = 1;\n")
	err := Run(context.Background(), Options{Path: path}, scriptRunner{bins: map[string]bool{"npx": true}})
	require.NoError(t, err)
	assert.Contains(t, out.String(), "properly formatted")
}

func TestRunCheckDirty(t *testing.T) {
	out, errOut := captureLog(t)
	path := writeTempFile(t, "const   x=1\n")
	t.Setenv("ENG_OXFMT_SCENARIO", "dirty")
	t.Setenv("ENG_OXFMT_DIRTY_FILE", path)
	err := Run(context.Background(), Options{Path: path}, scriptRunner{bins: map[string]bool{"npx": true}})
	require.ErrorIs(t, err, ErrUnformatted)
	assert.Contains(t, out.String(), path)
	assert.Contains(t, errOut.String(), "would be reformatted")
}

func TestRunCheckDirtyWithDiff(t *testing.T) {
	out, _ := captureLog(t)
	path := writeTempFile(t, "const   x=1\n")
	t.Setenv("ENG_OXFMT_SCENARIO", "dirty")
	t.Setenv("ENG_OXFMT_DIRTY_FILE", path)
	t.Setenv("ENG_OXFMT_FORMATTED", "const x = 1;\n")
	err := Run(
		context.Background(),
		Options{Path: path, Diff: true},
		scriptRunner{bins: map[string]bool{"npx": true}},
	)
	require.ErrorIs(t, err, ErrUnformatted)
	rendered := out.String()
	assert.Contains(t, rendered, "--- a/")
	assert.Contains(t, rendered, "-const   x=1")
	assert.Contains(t, rendered, "+const x = 1;")
}

func TestRunWrite(t *testing.T) {
	out, _ := captureLog(t)
	path := writeTempFile(t, "const   x=1\n")
	err := Run(
		context.Background(),
		Options{Path: path, Write: true},
		scriptRunner{bins: map[string]bool{"npx": true}},
	)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Formatted")
}

func TestRunValidationErrors(t *testing.T) {
	_, _ = captureLog(t)
	runner := scriptRunner{bins: map[string]bool{"npx": true}}
	path := writeTempFile(t, "const x = 1;\n")

	err := Run(context.Background(), Options{Path: path, Write: true, Check: true}, runner)
	assert.ErrorContains(t, err, "cannot combine --check and --write")

	err = Run(context.Background(), Options{Path: path, Write: true, Diff: true}, runner)
	assert.ErrorContains(t, err, "cannot combine --diff and --write")

	err = Run(context.Background(), Options{Path: path, Runner: "yarn"}, runner)
	assert.ErrorContains(t, err, `invalid --runner "yarn"`)
}

func TestRunMissingPath(t *testing.T) {
	_, _ = captureLog(t)
	err := Run(
		context.Background(),
		Options{Path: filepath.Join(t.TempDir(), "does-not-exist")},
		scriptRunner{bins: map[string]bool{"npx": true}},
	)
	assert.ErrorContains(t, err, "cannot access")
}

func TestRunNoRunnerAvailable(t *testing.T) {
	_, _ = captureLog(t)
	path := writeTempFile(t, "const x = 1;\n")
	err := Run(context.Background(), Options{Path: path}, scriptRunner{bins: map[string]bool{}})
	assert.ErrorContains(t, err, "no JS runner found")
}

func TestRunWarnsOnOldNode(t *testing.T) {
	_, errOut := captureLog(t)
	path := writeTempFile(t, "const x = 1;\n")
	t.Setenv("ENG_OXFMT_NODE_VERSION", "v20.11.0")
	err := Run(context.Background(), Options{Path: path}, scriptRunner{bins: map[string]bool{"npx": true}})
	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "node >= 22.18.0")
}
