package browser

import (
	"runtime"

	"github.com/eng618/eng/internal/execx"
)

var execCommand = execx.Command

// OpenURL opens the given URL in the system's default browser.
func OpenURL(url string) error {
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
