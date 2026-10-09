package compose

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui/theme"
)

var removeCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove a registered compose stack",
	Long: `Remove a user-registered compose stack from containers.stacks.

Discovered stacks under the legacy containers.path tree are not affected;
only explicitly added names can be removed.

With no arguments, an interactive picker lists the registered stacks.

Example:
  eng compose remove         # Interactive picker
  eng compose remove media`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) == 1 {
			return removeStack(args[0])
		}
		stacks := config.GetComposeStacks()
		if len(stacks) == 0 {
			log.Info("No registered compose stacks to remove.")
			return nil
		}
		idx, err := config.SelectComposeStack(stacks)
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				log.Info("Operation canceled.")
				return nil
			}
			return fmt.Errorf("failed to select compose stack: %w", err)
		}
		return removeStack(stacks[idx].Name)
	},
}

// removeStack removes one registered stack by name.
func removeStack(name string) error {
	if err := config.RemoveComposeStack(name); err != nil {
		return err
	}
	theme.SuccessMessage(fmt.Sprintf("Removed compose stack %q", name))
	return nil
}
