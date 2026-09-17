package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// ValidateCmd checks the resolved configuration and reports field errors.
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the resolved configuration",
	Long: `Loads configuration with ENG_ env overrides applied, runs central
validation, and reports one actionable error per invalid field.

Example: eng config validate`,
	RunE: func(_ *cobra.Command, _ []string) error {
		v := config.NewLoader(config.ResolveConfigPath(viper.ConfigFileUsed()))
		rc, err := config.LoadResolved(v)
		if err != nil {
			return err
		}
		errs := rc.Validate()
		if w := config.PlaintextSecretWarning(); w != nil {
			errs = append(errs, *w)
		}
		if len(errs) == 0 {
			theme.SuccessMessage("Configuration is valid")
			return nil
		}
		rows := make([][]string, 0, len(errs))
		for _, fe := range errs {
			rows = append(rows, []string{fe.Field, fe.Value, fe.Hint})
		}
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, ui.RenderTable(ui.TableOpts{
				Headers: []string{"FIELD", "VALUE", "FIX"},
				Rows:    rows,
			}))
		}
		theme.ErrorMessage("Configuration has validation errors")
		return nil
	},
}
