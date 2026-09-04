// Package version holds the application's build-time version metadata.
//
// It is intentionally a leaf package (no internal imports) so that both
// commands (cmd/version, cmd/doctor, cmd/root) and shared services
// (internal/telemetry) can read build info without layering inversions.
// The values are populated at build time via ldflags:
//
//	go build -ldflags="-X github.com/eng618/eng/internal/version.Version=…"
package version

// Build-time variables populated via ldflags (see .goreleaser.yaml and Taskfile.yaml).
var (
	// Version holds the application's version string (e.g., "0.1.0" or "dev").
	Version = "dev"
	// Commit holds the Git commit hash from which the application was built.
	Commit = "none"
	// Date holds the build date of the application.
	Date = "unknown"
)
