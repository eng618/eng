package update

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/eng618/eng/internal/sysinfo"
)

// Mockable system interaction points, migrated from cmd/system/system.go.
// Prefer internal/paths and internal/execx for new code.
var (
	execCommand  = exec.Command
	lookPath     = exec.LookPath
	userHomeDir  = os.UserHomeDir
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
