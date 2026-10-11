package oxfmt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/log"
)

// Run checks or formats target with Oxfmt using the @gv-tech/oxc-config style.
//
// Check mode (default) lists files that would change and returns
// ErrUnformatted when the workspace is dirty. Write mode formats in place.
func Run(ctx context.Context, opts Options, runner execx.Runner) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if runner == nil {
		runner = execx.Default
	}
	target := filepath.Clean(opts.Path)
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("cannot access %q: %w", opts.Path, err)
	}
	spec, err := ResolveLauncher(opts, runner.LookPath, hasBunLockFor(target))
	if err != nil {
		return err
	}
	if spec.Direct && opts.Runner == RunnerAuto && opts.ConfigPath == "" {
		log.Warn("No npx or bun runner found; using local oxfmt without the @gv-tech/oxc-config style")
	}
	warnOnOldNode(ctx, runner)
	log.Verbose(opts.IsVerbose, "Running oxfmt via %s on %s", spec.Binary, target)
	if opts.Write {
		return runWrite(ctx, runner, spec, opts, target)
	}
	return runCheck(ctx, runner, spec, opts, target)
}

// runWrite formats target in place, streaming oxfmt output to the log writers.
func runWrite(
	ctx context.Context,
	runner execx.Runner,
	spec LaunchSpec,
	opts Options,
	target string,
) error {
	cmd := runner.CommandContext(ctx, spec.Binary, invocationArgs(spec, opts, "--write", target)...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oxfmt --write failed: %w", err)
	}
	log.Success("Formatted %s with oxfmt (@gv-tech/oxc-config style)", target)
	return nil
}

// runCheck lists files that differ from oxfmt style without modifying them.
func runCheck(
	ctx context.Context,
	runner execx.Runner,
	spec LaunchSpec,
	opts Options,
	target string,
) error {
	cmd := runner.CommandContext(ctx, spec.Binary, invocationArgs(spec, opts, "--list-different", target)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		log.Success("All matched files are properly formatted (%s)", target)
		return nil
	}
	var exitErr *execx.ExitError
	if !errors.As(err, &exitErr) {
		return fmt.Errorf("oxfmt --list-different failed: %w", err)
	}
	files := parseListDifferent(string(out))
	if len(files) == 0 {
		detail := strings.TrimSpace(string(out) + "\n" + stderr.String())
		if detail != "" {
			return fmt.Errorf("oxfmt --list-different failed: %w\n%s", err, detail)
		}
		return fmt.Errorf("oxfmt --list-different failed: %w", err)
	}
	for _, f := range files {
		log.Message("  %s", f)
	}
	log.Warn("%d file(s) would be reformatted (run with --write to apply)", len(files))
	if opts.Diff {
		renderDiffs(ctx, runner, spec, opts, files)
	}
	return ErrUnformatted
}

// parseListDifferent extracts one file path per non-empty output line.
func parseListDifferent(out string) []string {
	var files []string
	for line := range strings.Lines(strings.TrimSpace(out)) {
		if path := strings.TrimSpace(line); path != "" {
			files = append(files, path)
		}
	}
	return files
}

// renderDiffs prints a colored unified-style diff per unformatted file,
// formatting each through oxfmt stdin mode so nothing on disk is touched.
func renderDiffs(
	ctx context.Context,
	runner execx.Runner,
	spec LaunchSpec,
	opts Options,
	files []string,
) {
	shown := 0
	for _, f := range files {
		if shown >= MaxDiffFiles {
			log.Info("... and %d more file(s) (diff truncated at %d)", len(files)-shown, MaxDiffFiles)
			return
		}
		orig, err := os.ReadFile(f)
		if err != nil {
			log.Warn("Skipping diff for %s: %v", f, err)
			continue
		}
		if len(orig) > MaxDiffBytes {
			log.Warn("Skipping diff for large file %s", f)
			continue
		}
		formatted, err := formatStdin(ctx, runner, spec, opts, f, orig)
		if err != nil {
			log.Warn("Skipping diff for %s: %v", f, err)
			continue
		}
		if string(formatted) == string(orig) {
			continue
		}
		fmt.Fprintln(log.Out, RenderDiff(f, string(orig), string(formatted)))
		shown++
	}
}

// formatStdin pipes input through oxfmt stdin mode and returns the result.
func formatStdin(
	ctx context.Context,
	runner execx.Runner,
	spec LaunchSpec,
	opts Options,
	name string,
	input []byte,
) ([]byte, error) {
	args := append([]string{}, spec.Args...)
	if opts.ConfigPath != "" {
		args = append(args, "-c", opts.ConfigPath)
	}
	args = append(args, "--stdin-filepath", name)
	cmd := runner.CommandContext(ctx, spec.Binary, args...)
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return nil, fmt.Errorf("oxfmt stdin formatting failed for %s: %s: %w", name, detail, err)
		}
		return nil, fmt.Errorf("oxfmt stdin formatting failed for %s: %w", name, err)
	}
	return stdout.Bytes(), nil
}
