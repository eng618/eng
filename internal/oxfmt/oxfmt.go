package oxfmt

import (
	"errors"
	"fmt"
	"strings"
)

// Runner selection values for Options.Runner.
const (
	RunnerAuto  = "auto"
	RunnerNpx   = "npx"
	RunnerBun   = "bun"
	RunnerOxfmt = "oxfmt"
)

// ConfigPackage is the shareable style package forwarded to the npx/bun wrapper.
const ConfigPackage = "@gv-tech/oxc-config@latest"

// MaxDiffFiles caps how many per-file diffs check mode renders with Options.Diff.
const MaxDiffFiles = 20

// MaxDiffBytes skips stdin formatting for files larger than this during --diff.
const MaxDiffBytes = 1 << 20

// ErrUnformatted is returned when check mode finds files needing formatting.
// Callers use errors.Is to distinguish "dirty workspace" from hard failures.
var ErrUnformatted = errors.New("formatting check failed: files would be reformatted")

// Options is the pure-data input for Run. The cmd layer adapts flags into it
// and never passes Viper/config values beyond plain fields.
type Options struct {
	Path       string
	Write      bool
	Check      bool
	Diff       bool
	Runner     string
	ConfigPath string
	IsVerbose  bool
}

// Validate rejects contradictory flag combinations and normalizes defaults.
func (o *Options) Validate() error {
	if o.Write && o.Check {
		return errors.New("cannot combine --check and --write: pick one mode")
	}
	if o.Write && o.Diff {
		return errors.New("cannot combine --diff and --write: --diff implies check mode")
	}
	if o.Runner == "" {
		o.Runner = RunnerAuto
	}
	o.Runner = strings.ToLower(o.Runner)
	switch o.Runner {
	case RunnerAuto, RunnerNpx, RunnerBun, RunnerOxfmt:
	default:
		return fmt.Errorf("invalid --runner %q: must be one of auto, npx, bun, oxfmt", o.Runner)
	}
	if o.Path == "" {
		o.Path = "."
	}
	return nil
}
