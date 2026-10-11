package oxfmt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/log"
)

// LaunchSpec describes the binary plus leading args used to invoke oxfmt.
// Direct specs call a local oxfmt; wrapper specs go through the gv-oxfmt
// runner (npx --package ... / bun x --package ...) which bundles the
// @gv-tech/oxc-config style.
type LaunchSpec struct {
	Binary string
	Args   []string
	Direct bool
}

// ResolveLauncher picks the oxfmt entrypoint for opts.Runner.
//
// Auto mode prefers the gv-oxfmt wrapper (bun, then npx) so the shared
// @gv-tech/oxc-config style applies even without a local install. A local
// oxfmt binary is only the last resort: it honors project config discovery
// (or ships defaults) rather than the shared style. With an explicit
// --config, a local oxfmt is preferred since no style package is needed.
//
// Callers may inject lookPath and bunLockExists for testing; pass nil/false
// for production behavior (execx.LookPath plus a bun.lock ancestor search).
func ResolveLauncher(opts Options, lookPath func(string) (string, error), bunLockExists bool) (LaunchSpec, error) {
	if lookPath == nil {
		lookPath = execx.LookPath
	}
	switch opts.Runner {
	case RunnerOxfmt:
		if _, err := lookPath("oxfmt"); err != nil {
			return LaunchSpec{}, fmt.Errorf("oxfmt not found on PATH: %w", err)
		}
		return LaunchSpec{Binary: "oxfmt", Direct: true}, nil
	case RunnerBun:
		if _, err := lookPath("bun"); err != nil {
			return LaunchSpec{}, fmt.Errorf("bun not found on PATH: %w", err)
		}
		return LaunchSpec{Binary: "bun", Args: wrapperArgs("bun")}, nil
	case RunnerNpx:
		if _, err := lookPath("npx"); err != nil {
			return LaunchSpec{}, fmt.Errorf("npx not found on PATH: %w", err)
		}
		return LaunchSpec{Binary: "npx", Args: wrapperArgs("npx")}, nil
	default:
		if opts.ConfigPath != "" {
			if _, err := lookPath("oxfmt"); err == nil {
				return LaunchSpec{Binary: "oxfmt", Direct: true}, nil
			}
		} else if bunLockExists {
			if _, err := lookPath("bun"); err == nil {
				return LaunchSpec{Binary: "bun", Args: wrapperArgs("bun")}, nil
			}
		}
		if _, err := lookPath("npx"); err == nil {
			return LaunchSpec{Binary: "npx", Args: wrapperArgs("npx")}, nil
		}
		if _, err := lookPath("bun"); err == nil {
			return LaunchSpec{Binary: "bun", Args: wrapperArgs("bun")}, nil
		}
		if _, err := lookPath("oxfmt"); err == nil {
			return LaunchSpec{Binary: "oxfmt", Direct: true}, nil
		}
		return LaunchSpec{}, fmt.Errorf(
			"no JS runner found: install oxfmt, or make bun or npx available on PATH",
		)
	}
}

// wrapperArgs returns the gv-oxfmt invocation prefix for npx or bun.
func wrapperArgs(runner string) []string {
	if runner == "bun" {
		return []string{"x", "--package", "oxfmt", "--package", ConfigPackage, "gv-oxfmt"}
	}
	return []string{"--yes", "--package", "oxfmt", "--package", ConfigPackage, "gv-oxfmt"}
}

// invocationArgs builds the full argv tail for a mode run (--write, --check,
// --list-different), honoring an explicit --config passthrough.
func invocationArgs(spec LaunchSpec, opts Options, mode, target string) []string {
	args := append([]string{}, spec.Args...)
	if opts.ConfigPath != "" {
		args = append(args, "-c", opts.ConfigPath)
	}
	return append(args, mode, target)
}

// hasBunLockFor reports whether target or one of its ancestors (up to the
// filesystem root) contains a bun.lock, mirroring package-manager detection.
func hasBunLockFor(target string) bool {
	abs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	dir := abs
	if info, err := os.Stat(abs); err == nil && !info.IsDir() {
		dir = filepath.Dir(abs)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "bun.lock")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}

// warnOnOldNode logs a best-effort warning when node predates the 22.18
// minimum required by @gv-tech/oxc-config TypeScript configs. It never fails.
func warnOnOldNode(ctx context.Context, runner execx.Runner) {
	cmd := runner.CommandContext(ctx, "node", "--version")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	ver := strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		return
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(strings.SplitN(parts[1], "-", 2)[0])
	if err1 != nil || err2 != nil {
		return
	}
	if major < 22 || (major == 22 && minor < 18) {
		log.Warn("Detected node %s; @gv-tech/oxc-config requires node >= 22.18.0", ver)
	}
}
