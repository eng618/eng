package update

import (
	"os"
	"runtime"

	"github.com/eng618/eng/internal/execx"
	"github.com/eng618/eng/internal/paths"
	"github.com/eng618/eng/internal/sysinfo"
)

// Mockable system interaction points, migrated from cmd/system/system.go.
// Seams default to the shared internal/execx and internal/paths entry
// points; tests override the package vars per-case as before.
var (
	execCommand  = execx.Command
	lookPath     = execx.LookPath
	userHomeDir  = paths.UserHomeDir
	stat         = os.Stat
	detectDistro = sysinfo.Detect
)

// openURL opens url in the default browser.
// TODO: move to internal/browser with setup/software_list.go openURL.
func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // linux, freebsd, openbsd, netbsd
		cmd = "xdg-open"
	}
	args = append(args, url)
	return execCommand(cmd, args...).Start()
}
