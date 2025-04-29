// /Users/EricGarciaMBP/Development/eng/cmd/version/version.go
package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set at build time using -ldflags
// They are exported so they can be used by root.go for the --version flag
var (
	Version = "dev"     // Default value if not built with ldflags
	Commit  = "none"    // Default value
	Date    = "unknown" // Default value
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of eng",
	Long:  `All software has versions. This is eng's. It shows the Git tag, commit hash, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("eng version: %s\n", Version)
		fmt.Printf("  Git Commit: %s\n", Commit)
		fmt.Printf("  Build Date: %s\n", Date)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
