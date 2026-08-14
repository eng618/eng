package system

import (
	"github.com/spf13/cobra"
)

var GPGCmd = &cobra.Command{
	Use:   "gpg",
	Short: "Manage GPG keys for commit signing and encryption",
	Long:  `Manage GPG keys for commit signing, encryption, initial setup, and expiration renewal.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return cmd.Help()
	},
}

func init() {
	GPGCmd.AddCommand(SetupGPGCmd)
	GPGCmd.AddCommand(RenewGPGCmd)
	GPGCmd.AddCommand(SyncGPGCmd)
}
