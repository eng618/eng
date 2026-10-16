package compose

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/paths"
	"github.com/eng618/eng/internal/ui/theme"
)

var addCmd = &cobra.Command{
	Use:   "add [name] [path]",
	Short: "Register a named compose stack at a local root path",
	Long: `Register a named compose stack pointing at an arbitrary local directory.

The compose file is autodetected inside the directory (docker-compose.yml/yaml,
compose.yml/yaml). The registration is stored under containers.stacks in the
config file and is included in list, status, up, down, pull, and logs.

With no arguments, an interactive wizard prompts for the name and path.

Example:
  eng compose add                  # Interactive wizard
  eng compose add media ~/Development/homelab/media`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		var name, rawPath string
		if len(args) == 2 {
			name, rawPath = args[0], args[1]
		} else {
			// Wizard flow: prefill with any partial args.
			initialName, initialPath := "", ""
			if len(args) == 1 {
				initialName = args[0]
			}
			var err error
			name, rawPath, err = config.PromptComposeStackValues(initialName, initialPath)
			if err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					log.Info("Operation canceled.")
					return nil
				}
				return err
			}
		}
		if err := registerStack(name, rawPath); err != nil {
			return err
		}
		return nil
	},
}

// registerStack validates a name/path pair and persists the registration.
func registerStack(name, rawPath string) error {
	if err := config.ValidateComposeStackName(name); err != nil {
		return err
	}
	absPath := paths.Expand(rawPath)
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("stack path %q: %w", rawPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("stack path %q is not a directory", rawPath)
	}
	composeFile, ok := containers.FindComposeFile(absPath)
	if !ok {
		return fmt.Errorf(
			"no compose file found in %q (looked for %v)",
			absPath,
			containers.ComposeFilenames,
		)
	}
	if err := config.AddComposeStack(config.ComposeStackEntry{Name: name, Path: absPath}); err != nil {
		return err
	}
	theme.SuccessMessage(fmt.Sprintf("Registered compose stack %q -> %s", name, absPath))
	log.Info("Compose file: %s", composeFile)
	return nil
}
