package telemetry

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/eng618/eng/internal/version"
)

// Event names
const (
	EventCommandExecuted = "cli_command_executed"
	EventDoctorRun       = "doctor_diagnostics_run"
	EventTestConnection  = "telemetry_connection_tested"
)

// IsCI returns true if the current environment is detected as a Continuous Integration runner.
func IsCI() bool {
	ciEnvVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"TRAVIS",
		"JENKINS_URL",
		"BITBUCKET_BUILD_NUMBER",
		"TEAMCITY_VERSION",
		"BUILDKITE",
	}

	for _, v := range ciEnvVars {
		if val := os.Getenv(v); val != "" && val != "0" && !strings.EqualFold(val, "false") {
			return true
		}
	}
	return false
}

// IsInteractive checks if standard output is connected to an interactive terminal TTY.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// DetectShell attempts to identify the current shell from the environment.
func DetectShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "unknown"
	}
	return filepath.Base(shellPath)
}

// ExtractCommandPath returns the sanitized command path (e.g. "git sync").
func ExtractCommandPath(cmd *cobra.Command) (string, string, string) {
	if cmd == nil {
		return "unknown", "unknown", ""
	}

	cmdPath := cmd.CommandPath()
	parts := strings.Fields(cmdPath)

	// Remove root command name "eng" if present at start
	if len(parts) > 0 && parts[0] == "eng" {
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return "root", "root", ""
	}

	rootCmd := parts[0]
	subCmd := ""
	if len(parts) > 1 {
		subCmd = strings.Join(parts[1:], " ")
	}

	return strings.Join(parts, " "), rootCmd, subCmd
}

// ExtractSanitizedFlags extracts ONLY the names of flags that were explicitly changed/passed,
// deliberately omitting flag values to prevent accidental leakage of sensitive tokens or paths.
func ExtractSanitizedFlags(cmd *cobra.Command) []string {
	if cmd == nil {
		return nil
	}

	var flags []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flags = append(flags, "--"+f.Name)
	})
	return flags
}

// CategorizeError categorizes an error without revealing private paths, tokens, or messages.
func CategorizeError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "unknown command"), strings.Contains(errStr, "unknown flag"):
		return "usage_error"
	case errors.Is(err, os.ErrNotExist):
		return "not_found"
	case errors.Is(err, os.ErrPermission):
		return "permission_denied"
	case strings.Contains(strings.ToLower(errStr), "context canceled"), strings.Contains(strings.ToLower(errStr), "signal: interrupt"):
		return "cancelled"
	case strings.Contains(strings.ToLower(errStr), "timeout"):
		return "timeout"
	default:
		return "execution_error"
	}
}

// BuildCommandProperties constructs a standardized, sanitized properties map for command execution.
func BuildCommandProperties(cmd *cobra.Command, args []string, duration time.Duration, execErr error) map[string]any {
	fullCmd, rootSub, subCmd := ExtractCommandPath(cmd)
	flags := ExtractSanitizedFlags(cmd)

	exitCode := 0
	if execErr != nil {
		exitCode = 1
	}

	props := map[string]any{
		"command":        fullCmd,
		"root_command":   rootSub,
		"subcommand":     subCmd,
		"duration_ms":    duration.Milliseconds(),
		"success":        execErr == nil,
		"exit_code":      exitCode,
		"error_category": CategorizeError(execErr),
		"flags_count":    len(flags),
		"flags":          flags,
		"args_count":     len(args),
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
		"cli_version":    version.Version,
		"build_commit":   version.Commit,
		"go_version":     runtime.Version(),
		"is_ci":          IsCI(),
		"is_interactive": IsInteractive(),
		"shell":          DetectShell(),
	}

	return props
}
