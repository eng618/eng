package auth

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils"
	"github.com/eng618/eng/internal/utils/log"
)

var (
	setTokenItem  string
	setToken      string
	setTokenStdin bool
	setHost       string
	setProject    string
	setNotes      string
)

// setCmd configures GitLab auth by saving a token to Bitwarden (optional) and storing references in config.
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Configure GitLab token and defaults",
	Long:  "Save or update a GitLab token in Bitwarden and set defaults (host, project, token item) in eng config.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If token is provided, we require a token-item name
		var tokenToSave string
		if setTokenStdin {
			in := bufio.NewScanner(os.Stdin)
			for in.Scan() {
				if tokenToSave != "" {
					tokenToSave += "\n"
				}
				tokenToSave += in.Text()
			}
			if err := in.Err(); err != nil {
				return fmt.Errorf("reading token from stdin failed: %w", err)
			}
		} else if setToken != "" {
			tokenToSave = setToken
		}

		if tokenToSave != "" {
			if strings.TrimSpace(setTokenItem) == "" {
				return errors.New("--token-item is required when providing a token")
			}
			if _, err := utils.SaveOrUpdateBitwardenSecret(
				setTokenItem,
				strings.TrimSpace(tokenToSave),
				setNotes,
			); err != nil {
				return err
			}
			log.Success("Saved/updated GitLab token in Bitwarden item: %s", setTokenItem)
		}

		// Persist config keys if provided
		changed := false
		if setTokenItem != "" {
			viper.Set("gitlab.tokenItem", setTokenItem)
			changed = true
		}
		if setHost != "" {
			viper.Set("gitlab.host", setHost)
			changed = true
		}
		if setProject != "" {
			viper.Set("gitlab.project", setProject)
			changed = true
		}
		if changed {
			if err := viper.WriteConfig(); err != nil {
				// If no config file yet, try SafeWriteConfig
				if os.IsNotExist(err) {
					if err := viper.SafeWriteConfig(); err != nil {
						return fmt.Errorf("failed writing config: %w", err)
					}
				} else {
					return fmt.Errorf("failed writing config: %w", err)
				}
			}
			log.Success("Updated eng config defaults")
		}

		if tokenToSave == "" && !changed {
			log.Warn("Nothing to do: provide --token/--stdin and/or defaults like --host/--project/--token-item")
		}
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(setCmd)
	setCmd.Flags().StringVar(&setTokenItem, "token-item", "", "Bitwarden item name to store/find the GitLab token")
	setCmd.Flags().StringVar(&setToken, "token", "", "GitLab token to save into Bitwarden (discouraged; prefer stdin)")
	setCmd.Flags().BoolVar(&setTokenStdin, "stdin", false, "Read token from STDIN")
	setCmd.Flags().StringVar(&setHost, "host", "", "Default GitLab host to save in config (e.g., gitlab.com)")
	setCmd.Flags().
		StringVar(&setProject, "project", "", "Default GitLab project path to save in config (e.g., group/sub/repo)")
	setCmd.Flags().StringVar(&setNotes, "notes", "", "Optional notes to save with the Bitwarden item")
}
