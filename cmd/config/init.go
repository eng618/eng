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

// InitCmd runs the guided configuration walkthrough.
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Guided walkthrough to configure eng step by step",
	Long: `Walks through profile, git, dotfiles, and telemetry settings one
section at a time, showing a preview before writing anything.

Resume mid-walkthrough with --from, or preview without writing with --dry-run.

Example: eng config init --from dotfiles --dry-run`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		fromStr, _ := cmd.Flags().GetString("from")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		step, err := config.ParseWizardStep(fromStr)
		if err != nil {
			return err
		}
		v := config.NewLoader(config.ResolveConfigPath(viper.ConfigFileUsed()))
		rc, err := config.LoadResolved(v)
		if err != nil {
			return err
		}
		changes, err := config.RunWizard(config.DefaultsFromResolved(rc), step)
		if err != nil {
			return err
		}
		rows := make([][]string, 0, len(changes))
		for k, val := range changes {
			rows = append(rows, []string{k, fmt.Sprintf("%v", val)})
		}
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, theme.InfoBox.Render("Planned changes:\n"+ui.RenderTable(ui.TableOpts{
				Headers: []string{"KEY", "VALUE"},
				Rows:    rows,
			})))
		}
		if dryRun {
			log.Info("Dry run: nothing written")
			return nil
		}
		for k, val := range changes {
			viper.Set(k, val)
		}
		path := config.ResolveConfigPath(viper.ConfigFileUsed())
		if err := viper.WriteConfigAs(path); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		theme.SuccessMessage(fmt.Sprintf("Configuration saved to %s", path))
		return nil
	},
}

func init() {
	InitCmd.Flags().String("from", "", "resume from step: profile, git, dotfiles, telemetry")
	InitCmd.Flags().Bool("dry-run", false, "preview changes without writing")
}
